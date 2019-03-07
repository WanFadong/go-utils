package model

import (
	"errors"
	"github.com/wanfadong/utils"
	"sync"

	"github.com/sirupsen/logrus"
)

/*
生产者-消费者模型 v2，不考虑断点的问题
*/

type Entry interface{}

type ProduceFunc func() (Entry, error)

type ConsumeFunc func(Entry) error

type ProducerConsumer struct {
	l *logrus.Entry

	produceFunc ProduceFunc
	consumeFunc ConsumeFunc
	consumerNum int

	buf  chan Entry
	fail chan struct{}
}

func NewProducerConsumer(produce ProduceFunc, consume ConsumeFunc, num int) (p *ProducerConsumer, err error) {

	if num <= 0 {
		err = errors.New("invalid consumer number")
		return
	}

	buf := make(chan Entry, num)
	fail := make(chan struct{})
	p = &ProducerConsumer{
		l:           logrus.WithField(utils.ReqidKey, "ProducerConsumer"),
		produceFunc: produce,
		consumeFunc: consume,
		consumerNum: num,

		buf:  buf,
		fail: fail,
	}
	return
}

// 查看缓冲区状态
func (p *ProducerConsumer) BufferStatus() int {
	return len(p.buf)
}

func (p *ProducerConsumer) Run() error {
	wg := sync.WaitGroup{}
	wg.Add(1 + p.consumerNum)

	var produceErr error
	go func() {
		defer wg.Done()
		produceErr = p.doProduce()
	}()

	consumeErrs := make([]error, p.consumerNum)
	for i := 0; i < p.consumerNum; i++ {
		go func(i int) {
			defer wg.Done()
			consumeErrs[i] = p.doConsume(i)
		}(i)
	}

	wg.Wait()

	if produceErr != nil {
		return produceErr
	}

	for _, consumeErr := range consumeErrs {
		if consumeErr != nil {
			return consumeErr
		}
	}

	return nil
}

func (p *ProducerConsumer) doProduce() error {
	for {
		entry, err := p.produceFunc()
		if err != nil {
			if err == ErrFinished {
				p.l.Info("produce finished, producer exit")
				close(p.buf)
				return nil
			} else {
				p.l.Error("produce failed, producer exit", err)
				closeChanSafely(p.fail)
				return err
			}
		}

		select {
		case <-p.fail:
			p.l.Info("producer/consumer failed, producer exit")
			return nil
		case p.buf <- entry:
			p.l.Debug("produce entry", entry)
		}
	}
}

func (p *ProducerConsumer) doConsume(index int) error {
	for {
		select {
		case <-p.fail:
			p.l.Infof("producer/consumer failed, consumer %v exit", index)
			return nil
		case entry, ok := <-p.buf:
			if !ok {
				p.l.Infof("produce finished, consumer %v exit", index)
				return nil
			}

			err := p.consumeFunc(entry)
			if err != nil {
				p.l.Errorf("consume failed, consumer %v exit, err: %v", index, err)
				closeChanSafely(p.fail)
				return err
			}
		}
	}
}

func closeChanSafely(c chan struct{}) {
	select {
	case <-c:
	default:
		close(c)
	}
}
