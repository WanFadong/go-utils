/*
Package model have some util model(模型).
*/
package model

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"github.com/wanfadong/utils"
	"github.com/wanfadong/utils/tool"
	xlog "github.com/sirupsen/logrus"
)

// ErrFinished is an error flag that indicates the end of producing
var ErrFinished = errors.New("produce finished")

// E is a util interface that Produce results must implement
type E interface {
	String() string
	Marker() int64     // 用于记录失败时的位置
	NextMarker() int64 // produce 出错时，只能拿到上一个处理的条目，所以需要记录 nextMarker
}

// Producer Produce E
type Producer interface {
	Produce() (E, error)
}

// Consumer consume E
type Consumer interface {
	Consume(E) error
}

// ProduceConsumer Produce&consume E
type ProduceConsumer interface {
	Producer
	Consumer
}

// ProducerConsumerConfig is the config of a pcr
type ProducerConsumerConfig struct {
	Produce        Producer
	Consume        Consumer
	OutStop        chan struct{}
	Num            int
	ScCfg          tool.SpeedometerConfig
	MarkerFilePath string
}

// ProducerConsumerRunner is a realized Producer-Consumer model
type ProducerConsumerRunner struct {
	xl           *xlog.Logger
	producer     Producer
	consumer     Consumer
	num          int           // 消费者数量
	buf          chan E        // 缓冲区。
	stop         chan struct{} // 避免 consume 直接关闭 buf chan 导致 panic
	outStop      chan struct{} // 给外部提供停止的入口
	speedCounter *tool.Speedometer

	m              sync.Mutex
	marker         int64 // 处理过程出问题时，这个值表示第一个没有处理entry的位置
	markerFilePath string
}

// NewProducerConsumerRunner return a ProducerConsumerRunner instance with given config
// todo 提供使用方法的构造方法
func NewProducerConsumerRunner(xl *xlog.Logger, cfg ProducerConsumerConfig) (p *ProducerConsumerRunner, err error) {
	if cfg.Num <= 0 {
		err = errors.New("invalid consumer num, < 0")
		xl.Error(err)
		return
	}
	buf := make(chan E, cfg.Num)
	stop := make(chan struct{})
	speedCounter := tool.NewSpeedometer(xl, cfg.ScCfg)
	p = &ProducerConsumerRunner{
		xl:             xl,
		producer:       cfg.Produce,
		consumer:       cfg.Consume,
		num:            cfg.Num,
		buf:            buf,
		stop:           stop,
		speedCounter:   speedCounter,
		outStop:        cfg.OutStop,
		markerFilePath: cfg.MarkerFilePath,
	}
	return
}

// Run run a Produce-consume task, with given consume and Produce func.
// 除非所有的 consumer 都失败了，否则一定会把 Produce 出来的 entry 全部处理完成后才退出。
// 断点处理：
// 1. produce失败，producer 会记录失败的位置。consume会处理已经produce的数据，然后退出。
// 2. consume失败，consumer 会记录失败的位置。producer 立即退出，并记录结束的位置。其他 consumer 会继续处理 buf 中剩余的数据。
// 结束的位置：
// 	producer：处理完成的最后一条
// 	consumer：处理失败的那条数据。
func (p *ProducerConsumerRunner) Run() {
	wg := sync.WaitGroup{}
	wg.Add(1 + p.num)

	go func() {
		defer wg.Done()
		p.doProduce()
	}()

	for i := 0; i < p.num; i++ {
		go func(i int) {
			defer wg.Done()
			p.doConsume(i)
		}(i)
	}

	wg.Wait()

	if p.marker != 0 {
		p.recordMarker()
	}
	// 统计处理的结果
	err := p.speedCounter.Close()
	if err != nil {
		p.xl.Error("speed counter close failed", err)
	}
}

func (p *ProducerConsumerRunner) doProduce() {
	xl := p.xl

	var lastCommitEntry E
	for {
		// todo-是不是先检查更好？
		entry, err := p.producer.Produce()
		if err != nil {
			if err == ErrFinished {
				xl.Info("producer exit because of finished")
			} else {
				xl.Info("producer exit because of Produce err:", err)
				p.setMarker(lastCommitEntry, true)
			}
			close(p.buf)
			return
		}

		select {
		case <-p.outStop:
			xl.Info("producer exit because of out stop")
			p.setMarker(lastCommitEntry, true)
			close(p.buf)
			return
		case <-p.stop:
			xl.Info("producer exit because of consumer err")
			p.setMarker(lastCommitEntry, true)
			close(p.buf)
			return
		case p.buf <- entry:
			p.speedCounter.Increase()
		}

		lastCommitEntry = entry
	}
}

func (p *ProducerConsumerRunner) doConsume(i int) {
	xl := p.xl

	for {
		select {
		case entry, ok := <-p.buf:
			// 结束
			if !ok {
				xl.Info("consumer exit because of finished or producer err, consumer index: ", i)
				return
			}
			err := p.consumer.Consume(entry)
			if err != nil {
				xl.Infof("consumer exit because of err, consumer index: %v, err: %v", i, err)
				p.setMarker(entry, false)
				safeClose(p.stop) // consumer 关闭 stop chan，而不要关闭 buf chan，避免 panic。
				return
			}
		}
	}
}

// 如果需要记录到文件中，那么选取最小的 marker
func (p *ProducerConsumerRunner) setMarker(entry E, next bool) {
	xl := p.xl
	if entry != nil {
		xl.Infof("marker: %v, next marker: %v, use next marker: %v", entry.Marker(), entry.NextMarker(), next)
		p.m.Lock()
		defer p.m.Unlock()

		thisMarker := entry.Marker()
		if next {
			thisMarker = entry.NextMarker()
		}
		if p.marker == 0 || thisMarker < p.marker {
			p.marker = thisMarker
		}
	}
}

func (p *ProducerConsumerRunner) recordMarker() {

	p.xl.Infof("final marker is %v", p.marker)
	if p.markerFilePath != "" {
		exists, err := utils.IsFileExists(p.markerFilePath)
		if err != nil {
			p.xl.Error("get file exists failed", err)
			return
		}
		if exists {
			err = os.Rename(p.markerFilePath, p.markerFilePath+".backup")
			if err != nil {
				p.xl.Error("rename file failed", p.markerFilePath, err)
				// 出错也会继续
			}
		}
		file, err := os.Create(p.markerFilePath)
		if err != nil {
			p.xl.Error("create file failed", p.markerFilePath, err)
			return
		}
		_, err = file.WriteString(strconv.FormatInt(p.marker, 10))
		if err != nil {
			p.xl.Error("write string failed", p.marker, err)
			return
		}
	}
}

func safeClose(stop chan struct{}) {
	select {
	case <-stop: // 已经被关闭
	default:
		close(stop)
	}
}
