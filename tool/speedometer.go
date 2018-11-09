package tool

import (
	"strconv"
	"time"

	"github.com/qiniu/xlog.v1"
)

// SpeedometerConfig is Speedometer Config
type SpeedometerConfig struct {
	Total                    int64 `json:"total"`
	ProcessedBefore          int64 `json:"processed_before"`
	OutputTimeIntervalSecond int64 `json:"output_time_interval_second"`
	OutputNumInterval        int64 `json:"output_num_interval"`
}

// Speedometer is a util tool for counting speed and processingStatics statics regularly
// 处理数量/时间，速度单位是：个/s，精确到个位数
// 每隔多长时间；每隔多少数量输出一次当前处理信息
// 支持续处理
type Speedometer struct {
	xl                       *xlog.Logger
	startTime                int64
	outputTimeIntervalSecond int64 // s
	lastOutputTime           int64
	outputNumInterval        int64
	lastOutputNum            int64
	total                    int64 // 可以是估计值
	processed                int64 // 这次处理的
	processedBefore          int64 // 之前处理的
}

// NewSimpleSpeedometer return the most simple sc.
func NewSimpleSpeedometer(xl *xlog.Logger, timeIntervalSecond int64) (s *Speedometer) {

	s = &Speedometer{
		xl:                       xl,
		startTime:                time.Now().Unix(),
		lastOutputTime:           time.Now().Unix(),
		outputTimeIntervalSecond: timeIntervalSecond,
	}
	return
}

// NewSpeedometer is a constructor
func NewSpeedometer(xl *xlog.Logger, cfg SpeedometerConfig) (s *Speedometer) {

	s = &Speedometer{
		xl:                       xl,
		startTime:                time.Now().Unix(),
		total:                    cfg.Total,
		outputTimeIntervalSecond: cfg.OutputTimeIntervalSecond,
		lastOutputTime:           time.Now().Unix(),
		outputNumInterval:        cfg.OutputNumInterval,
		processedBefore:          cfg.ProcessedBefore,
	}
	return
}

func output(xl *xlog.Logger, msg string) {
	xl.Info(msg)
}

// basic + 预计剩余时间(-h)
// todo: 最近一分钟速度
func (s *Speedometer) processingStatics() {

	var msg string
	msg += "processing..."

	msg += join("processed", strconv.FormatInt(s.processed, 10))
	now := time.Now().Unix()

	usedTime := now - s.startTime
	if usedTime != 0 {
		speed := s.processed / usedTime
		msg += join("speed", strconv.FormatInt(speed, 10))
		if s.total != 0 {
			leftNum := s.total - s.processed - s.processedBefore
			s.xl.Debug(s.processed, usedTime, leftNum)
			leftTime := leftNum * usedTime / s.processed
			leftHumanTime := humanTime(leftTime)
			msg += join("left time", leftHumanTime)
		}
	}
	output(s.xl, msg)
}

// 名称，总处理量，总共用时，平均速度。
func (s *Speedometer) overallStatics() {
	var msg string
	msg += "finished..."
	msg += join("total", strconv.FormatInt(s.processed+s.processedBefore, 10))
	msg += join("processed this", strconv.FormatInt(s.processed, 10))
	msg += join("processed before", strconv.FormatInt(s.processedBefore, 10))
	msg += join("processed total", strconv.FormatInt(s.processed+s.processedBefore, 10))
	now := time.Now().Unix()
	t := now - s.startTime
	msg += join("use time", humanTime(t))

	if t != 0 {
		speed := s.processed / t // ns 不可能为0
		msg += join("speed", strconv.FormatInt(speed, 10))
	}
	output(s.xl, msg)
}

func (s *Speedometer) outputCheck() {
	// 时间和数量分开检测。
	if s.outputNumInterval != 0 && s.processed-s.lastOutputNum >= s.outputNumInterval {
		s.lastOutputNum = s.processed
		s.processingStatics()
	}
	if s.outputTimeIntervalSecond != 0 && time.Now().Unix()-s.lastOutputTime >= s.outputTimeIntervalSecond {
		s.lastOutputTime = time.Now().Unix()
		s.processingStatics()
	}
}

// Start 开始计时。否则，会使用 New 的时间作为开始时间。
func (s *Speedometer) Start() {
	s.startTime = time.Now().Unix()
}

// Increase add 1
func (s *Speedometer) Increase() {
	s.processed++
	s.outputCheck()
}

// IncreaseN add n
func (s *Speedometer) IncreaseN(n int) {
	s.processed += int64(n)
	s.outputCheck()
}

// Close show overall statics and record processed to file
func (s *Speedometer) Close() (err error) {
	s.overallStatics()
	return
}

func humanTime(t int64) string {
	var head string
	if t < 0 {
		t = -t
		head = "-"
	}
	names := []string{"m", "h", "d"}
	ratios := []int64{60, 60, 24}
	name := "s"
	for i, r := range ratios {
		if t < r {
			break
		}
		t = t / r
		name = names[i]
	}
	return head + strconv.FormatInt(t, 10) + name
}

func join(key string, val string) (s string) {
	s = key + ": " + val + ", "
	return
}
