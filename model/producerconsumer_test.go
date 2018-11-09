package model

import (
	"strconv"
	"testing"

	"path/filepath"

	"github.com/golib/assert"
	"github.com/qiniu/xlog.v1"
	"qbox.us/errors"
	"somewan.com/utils/tool"
)

var (
	i         int64
	outputDir = "/Users/wan/workspace/go/data/output/"
)

type simpleEntry struct {
	id int64
}

func (s simpleEntry) String() string {
	return strconv.FormatInt(s.id, 10)
}

func (s simpleEntry) Marker() int64 {
	return s.id
}

func (s simpleEntry) NextMarker() int64 {
	return s.id + 1
}

//
func TestProducerConsumer_Run(t *testing.T) {
	i = 0
	cfg := ProducerConsumerConfig{
		Produce: &ProduceOk{},
		Consume: &ConsumeOk{},
		Num:     2,
		ScCfg:   tool.SpeedometerConfig{},
	}
	p, err := NewProducerConsumerRunner(xlog.NewDummy(), cfg)
	assert.NoError(t, err)
	p.Run()
}

func TestProducerConsumer_Run_ProduceErr(t *testing.T) {
	i = 0
	cfg := ProducerConsumerConfig{
		Produce: &ProduceBad{},
		Consume: &ConsumeOk{},
		Num:     2,
		ScCfg:   tool.SpeedometerConfig{},
	}
	p, err := NewProducerConsumerRunner(xlog.NewDummy(), cfg)
	assert.NoError(t, err)
	p.Run()
}

func TestProducerConsumer_Run_ConsumeErr(t *testing.T) {
	i = 0
	cfg := ProducerConsumerConfig{
		Produce: &ProduceOk{},
		Consume: &ConsumeBad{},
		Num:     2,
		ScCfg:   tool.SpeedometerConfig{},
	}
	p, err := NewProducerConsumerRunner(xlog.NewDummy(), cfg)
	assert.NoError(t, err)
	p.Run()
}

// 断点自动续处理
func TestProducerConsumerRunner_Run_Marker(t *testing.T) {
	i = 0

	path := filepath.Join(outputDir, "test-marker.txt")
	// 处理20个之后失败
	cfg := ProducerConsumerConfig{
		Produce:        &ProduceOk{},
		Consume:        &ConsumeBad{},
		Num:            2,
		ScCfg:          tool.SpeedometerConfig{},
		MarkerFilePath: path,
	}
	p, err := NewProducerConsumerRunner(xlog.NewDummy(), cfg)
	assert.NoError(t, err)
	p.Run()

	// 续处理剩下的20个
	cfg2 := ProducerConsumerConfig{
		Produce:        &ProduceOk{},
		Consume:        &ConsumeOk{},
		Num:            2,
		ScCfg:          tool.SpeedometerConfig{},
		MarkerFilePath: path,
	}
	p2, err := NewProducerConsumerRunner(xlog.NewDummy(), cfg2)
	assert.NoError(t, err)
	p2.Run()
}

type ProduceOk struct {
}

type ConsumeOk struct {
}

type ProduceBad struct {
}

type ConsumeBad struct {
}

func (p *ProduceOk) Produce() (entry E, err error) {
	i++
	if i <= 40 {
		entry = simpleEntry{id: i}
	} else {
		err = ErrFinished
	}
	return
}

func (p *ProduceBad) Produce() (entry E, err error) {
	i++
	if i <= 20 {
		entry = simpleEntry{id: i}
	} else {
		err = errors.New("producer exceed")
	}
	return
}

func (c *ConsumeOk) Consume(entry E) (err error) {
	xl := xlog.NewWith("Consume")
	xl.Info(entry.String())
	return
}

func (c *ConsumeBad) Consume(entry E) (err error) {
	xl := xlog.NewWith("ConsumeWithError")
	s := entry.String()
	id, _ := strconv.Atoi(s)
	if id <= 20 {
		xl.Info(s)
	} else {
		err = errors.New("consumer exceed")
	}
	return
}
