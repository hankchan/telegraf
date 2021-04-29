package gbt32960

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	ErrNoMetric = fmt.Errorf("no metric in line")
)

// Parser decodes logfmt formatted messages into metrics.
type Parser struct {
	MetricName  string
	DefaultTags map[string]string
	Now         func() time.Time
}

// NewParser creates a parser.
func NewParser(metricName string, defaultTags map[string]string) *Parser {
	return &Parser{
		MetricName:  metricName,
		DefaultTags: defaultTags,
		Now:         time.Now,
	}
}

// Parse converts a slice of bytes in logfmt format to metrics.
func (p *Parser) Parse(b []byte) ([]telegraf.Metric, error) {

	var msg = GBT32960Message{}

	offset := 0
	if b[0] != 0x20 {
		// 默认情况，kafka: iccid(20)+datatime(19)+gbt32960()
		offset = 39
		msg.iccid = string(b[0:20])          // kafka里国标数据前有iccid
		msg.strServerTime = string(b[20:39]) // kafka里国标数据前有格式化的时间戳
	} else {
		// 异常情况，有时kafka会出现第一个是0x20(空格)的情况，即没有iccid
		// unknown msg start: 20323032312d30342d32392031313a31363a3339232303fe4c45575445423134344
		offset = 20
		msg.iccid = ""
		msg.strServerTime = string(b[1:20]) // 异常情况时，只有时间戳
	}

	// 获取服务器时间戳
	loc, _ := time.LoadLocation("Asia/Shanghai")
	msg.serverTime, _ = time.ParseInLocation("2006-01-02 15:04:05", msg.strServerTime, loc)

	// GBT 32960 报文头解析
	// GBT 32960, [0,1]，起止符
	if b[offset+0] != 0x23 || b[offset+1] != 0x23 {
		log.Panicf("-> TOOD: msg start err: %x...", b)
		return nil, nil
	}

	msg.cmd = b[offset+2]
	msg.resp = b[offset+3]
	msg.vin = b[offset+4 : offset+21]
	msg.len = uint16(b[offset+22])<<8 | uint16(b[offset+23]) + 25 // 消息封装固定长度为25
	msg.body = b[offset : int(msg.len)+offset-1]

	//log.Printf("-> msg: %+v %x\n", msg, b)

	if msg.cmd == 0x01 || msg.cmd == 0x02 || msg.cmd == 0x03 || msg.cmd == 0x04 {
		//log.Printf("-> time: %d %d %d %d %d %d %v \n", 2000+int(b[offset+24]), int(b[offset+25]), int(b[offset+26]), int(b[offset+27]), int(b[offset+28]), int(b[offset+29]), b)
		msg.tboxTime = time.Date(2000+int(b[offset+24]), time.Month(int(b[offset+25])), int(b[offset+26]), int(b[offset+27]), int(b[offset+28]), int(b[offset+29]), 0, loc)
	}

	var xor uint8
	for i := 2; i < len(msg.body)-1; i++ {
		xor ^= msg.body[i]
	}

	// TODO: check crc

	//log.Printf("-> msg: %v %v %s %s %d %d\n", msg.serverTime, msg.tboxTime, string(msg.vin), string(msg.iccid), msg.len, len(msg.body))

	// GBT 32960 报文体解析

	var err error
	var mapResult map[string]interface{}
	var GBT32960 = GBT32960Protocol{}

	switch {
	case msg.cmd == 0x01:
		// 车辆登入
		if mapResult, err = GBT32960.UnpackEVLogin(&msg); err != nil {
			log.Panicf("-> TOOD: UnpackEVLogin err: %x...", b)
			return nil, ErrNoMetric
		}
	case msg.cmd == 0x02 || msg.cmd == 0x03:
		// 实时信息上报 or 补发信息上报
		if mapResult, err = GBT32960.UnpackEVData(&msg); err != nil {
			log.Panicf("-> TOOD: UnpackEVData err: %x...", b)
			return nil, ErrNoMetric
		}
	case msg.cmd == 0x04:
		// 车辆登出
		if mapResult, err = GBT32960.UnpackEVLogout(&msg); err != nil {
			log.Panicf("-> TOOD: UnpackEVLogout err: %x...", b)
			return nil, ErrNoMetric
		}
	default:
		// TODO: 暂不处理其他命令类型,参考CheckCommandFlag函数注释
		//log.Printf("-> TOOD: msg.cmd err: %x %s %x...", msg.cmd, GBT32960.CheckCommandFlag(msg.cmd), b)
		return nil, nil
	}

	// GBT 32960 报文解析结果存入influxDB
	if mapResult != nil {
		metrics := make([]telegraf.Metric, 0)
		// m := metric.New("gbt32960", map[string]string{"vin": string(msg.vin)}, mapResult, strTime)
		m := metric.New("gbt32960", map[string]string{"vin": string(msg.vin), "iccid": string(msg.iccid)}, mapResult, msg.tboxTime)
		metrics = append(metrics, m)
		return metrics, nil
	} else {
		return nil, nil
	}
}

// ParseLine converts a single line of text in logfmt format to metrics.
func (p *Parser) ParseLine(s string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(s))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}
	return metrics[0], nil
}

// SetDefaultTags adds tags to the metrics outputs of Parse and ParseLine.
func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
