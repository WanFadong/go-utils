package tool

import (
	"fmt"
	"testing"
	"time"

	"os"

	xlog "github.com/sirupsen/logrus"
)

var (
	outputDir = "/Users/wan/workspace/go/data/output/"
)

func init() {
	xlog.SetOutputLevel(0)
}

func TestSpeedCounter_TimeOutput(t *testing.T) {
	processedPath := outputDir + "processed.txt"
	defer os.Remove(processedPath)

	scCfg := SpeedometerConfig{
		Total:                    500,
		OutputTimeIntervalSecond: 1,
	}
	counter := NewSpeedometer(xlog.NewDummy(), scCfg)
	for i := 0; i < 400; i++ {
		time.Sleep(time.Millisecond * 10)
		counter.Increase()
	}
	counter.Close()

	// 续处理
	counter2 := NewSpeedometer(xlog.NewDummy(), scCfg)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 10)
		counter2.Increase()
	}
	counter2.Close()

}

func TestSpeedCounter_NumOutput(t *testing.T) {
	scCfg := SpeedometerConfig{
		OutputNumInterval: 100,
	}
	counter := NewSpeedometer(xlog.NewDummy(), scCfg)
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * 10)
		counter.Increase()
	}
	counter.Close()
}

func TestHumanTime(t *testing.T) {
	tests := []int64{60 * 60 * 24, 3, 3}
	for _, test := range tests {
		ht := humanTime(test)
		fmt.Println(ht)
	}
}
