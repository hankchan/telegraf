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

	offset := 39 // kafka: iccid(20)+datatime(19)+gbt32960()

	//log.Printf("buf: %x \n", b[offset:])

	// GBT 32960, [0,1]，起止符
	if b[offset+0] != 0x23 || b[offset+1] != 0x23 {
		log.Printf("-> TOOD: unknown msg start: %x...", b[:8])
		return nil, ErrNoMetric
	}

	msg := GBT32960Message{
		iccid:   string(b[0:20]),  // kafka里国标数据前有iccid
		strTime: string(b[20:39]), // kafka里国标数据前有格式化的时间戳

		cmd:  b[offset+2],
		resp: b[offset+3],
		vin:  b[offset+4 : offset+21],
		// skip BT 32960, [21], 数据单元加密方式
	}

	msg.len = uint16(b[offset+22])<<8 | uint16(b[offset+23]) + 25 // 消息封装固定长度为25
	msg.body = b[offset : int(msg.len)+offset-1]

	var xor uint8
	for i := 2; i < len(msg.body)-1; i++ {
		xor ^= msg.body[i]
	}
	// TODO: chech crc

	//log.Printf("msg: %s %s %s %d %d\n", string(datetime), string(iccid), string(msg.vin), msg_len, len(msg.body))

	var mapResult map[string]interface{}
	var GBT32960 = GBT32960Protocol{}

	switch {
	case msg.cmd == 0x01: // 车辆登入
		if err := GBT32960.UnpackEVLogin(&msg, &mapResult); err != nil {
			return nil, ErrNoMetric
		}
	case msg.cmd == 0x02: // 实时信息上报
	case msg.cmd == 0x03: // 补发信息上报
		if err := GBT32960.UnpackEVData(&msg, &mapResult); err != nil {
			return nil, ErrNoMetric
		}
	case msg.cmd == 0x04: // 车辆登出
		if err := GBT32960.UnpackEVLogout(&msg, &mapResult); err != nil {
			return nil, ErrNoMetric
		}
	}

	// if metric, err := metric.New("gbt32960", map[string]string{"iccid": msg.iccid, "vin": string(msg.vin)}, mapResult, time.Now().UTC()); err != nil {
	// 	return nil, err
	// } else {
	// 	return []telegraf.Metric{metric}, nil
	// }

	return metric.New("gbt32960", map[string]string{"iccid": msg.iccid, "vin": string(msg.vin)}, mapResult, time.Now().UTC()), nil
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
