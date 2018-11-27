package model

import (
	"sync"

	"github.com/qiniu/xlog.v1"
)

/*
生产者-消费者模型 v2，不考虑断点的问题
*/

type Entry interface{}

type ProduceFunc func() (Entry, error)

type ConsumeFunc func(Entry) error

type ProducerConsumer struct {
	xl *xlog.Logger

	produceFunc ProduceFunc
	consumeFunc ConsumeFunc
	consumerNum int

	buf  chan Entry
	fail chan struct{}
}

func NewProducerConsumer(produce ProduceFunc, consume ConsumeFunc, num int) (p *ProducerConsumer, err error) {

	buf := make(chan Entry, num)
	fail := make(chan struct{})
	p = &ProducerConsumer{
		xl:          xlog.NewWith("ProducerConsumer"),
		produceFunc: produce,
		consumeFunc: consume,
		consumerNum: num,

		buf:  buf,
		fail: fail,
	}
	return
}

func (p *ProducerConsumer) Run() {
	wg := sync.WaitGroup{}
	wg.Add(1 + p.consumerNum)

	go func() {
		defer wg.Done()
		p.doProduce()
	}()

	for i := 0; i < p.consumerNum; i++ {
		go func(i int) {
			defer wg.Done()
			p.doConsume(i)
		}(i)
	}

	wg.Wait()

}

func (p *ProducerConsumer) doProduce() error {
	for {
		entry, err := p.produceFunc()
		if err != nil {
			if err == ErrFinished {
				p.xl.Info("produce finished, producer exit")
				close(p.buf)
			} else {
				p.xl.Error("produce failed, producer exit", err)
				closeChanSafely(p.fail)
			}
			return err
		}

		select {
		case <-p.fail:
			p.xl.Info("producer/consumer failed, producer exit")
			return nil
		case p.buf <- entry:
			p.xl.Debugf("produce entry", entry)
		}
	}
}

func (p *ProducerConsumer) doConsume(index int) error {
	for {
		select {
		case <-p.fail:
			p.xl.Infof("producer/consumer failed, consumer %v exit", index)
			return nil
		case entry, ok := <-p.buf:
			if !ok {
				p.xl.Infof("produce finished, consumer %v exit", index)
				return nil
			}

			err := p.consumeFunc(entry)
			if err != nil {
				p.xl.Errorf("consume failed, consumer %v exit, err: %v", index, err)
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
