package gbt32960

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	ErrGbt32960Protocol = fmt.Errorf("error gbt 32960 protocol")
)

// GBT32960Message struct
type GBT32960Message struct {
	// kafka 消息头
	iccid         string
	strServerTime string

	// 时间戳
	tboxTime   time.Time
	serverTime time.Time

	// 国标数据结构
	vin  []byte
	len  uint16
	cmd  byte
	resp byte
	body []byte
}

// GBT32960Protocol struct
type GBT32960Protocol struct {
}

const (
	DefCfg_Cell_Num   = 41
	DefCfg_Temp_Num   = 15
	DefCfg_Subsys_Num = 1
	DefCfg_Motor_Num  = 1
	/*max fault number is 50*/
	MAX_FAULT_NUM = 35
	/*if more than 200 must split package*/
	MAX_NUM_PACKAGE = 200
	/*pure data size*/
	MAX_DATA_SIZE = 2000 /*pure data size*/
)

type EvData struct {
	Sn            uint16 `json:"登入流水号"`
	Iccid         string `json:"ICCID"`
	Pack_count    byte   `json:"可充电储能子系统数"`
	Pack_code_len byte   `json:"可充电储能系统编码长度"`
	Pack_code     string `json:"可充电储能系统编码"`
}

type EvLogout struct {
	Sn uint16 `json:"登出流水号"`
}

type VcuData struct {
	//command         uint8
	Vehicle_status  uint8  `json:"车辆状态"`
	Charging_status uint8  `json:"充电状态"`
	Operation_mode  uint8  `json:"运行模式"`
	Speed           uint16 `json:"车速"`
	Mileage         uint32 `json:"累计里程"`
	Voltage         uint16 `json:"总电压"`
	Current         uint16 `json:"总电流"`
	Soc             uint8  `json:"SOC"`
	Dc_status       uint8  `json:"DC-DC状态"`
	Stall           uint8  `json:"档位"`
	Resistance      uint16 `json:"绝缘电阻"`
	Acc_pedal       uint8  `json:"加速踏板行程值"`
	Braking_status  uint8  `json:"制动踏板状态"`
}

/*
 * Driving motor data
 */
type MotorInfo struct {
	Motor_SN     uint8  `json:"驱动电机序号"`
	Motor_status uint8  `json:"驱动电机状态"`
	Motor_ctr    uint8  `json:"驱动电机控制器温度"`
	Motor_speed  uint16 `json:"驱动电动转速"`
	Motororque   uint16 `json:"驱动电机转矩"`
	Motor        uint8  `json:"驱动电机温度"`
	Motor_ctr_v  uint16 `json:"电机控制器输入电压"`
	Motor_ctr_c  uint16 `json:"电机控制器直流母线电流"`
}

type MotorData struct {
	motor_sum uint8
	motors    [DefCfg_Motor_Num]MotorInfo
}

/*gps data */
type PositionData struct {
	Loc_status uint8  `json:"定位状态"`
	Longitude  uint32 `json:"经度"`
	Latitude   uint32 `json:"纬度"`
}

/*extreme value data*/
type ExtrmvData struct {
	Hv_subsys_num uint8  `json:"最高电压电池子系统号"`
	Hv_cell_num   uint8  `json:"最高电压电池单体代号"`
	Shv           uint16 `json:"电池单体电压最高值"`
	Lv_subsys_num uint8  `json:"最低电压电池子系统号"`
	Lv_cell_num   uint8  `json:"最低电压电池单体代号"`
	Slv           uint16 `json:"电池单体电压最低值"`
	Ht_subsys_num uint8  `json:"最高温度子系统号"`
	Ht_cell_num   uint8  `json:"最高温度探针序号"`
	Sht           uint8  `json:"最高温度值"`
	Lt_subsys_num uint8  `json:"最低温度子系统号"`
	Lt_cell_num   uint8  `json:"最低温度探针序号"`
	Slt           uint8  `json:"最低温度值"`
}

/*alarm data*/
type AlarmData struct {
	Highst_alart_level  uint8    `json:"最高报警等级"`
	General_alarm_flag  uint32   `json:"通用报警标志"`
	Battery_alarm_sum   uint8    `json:"可充电储能装置故障总数N1"`
	Battery_alarm_list  []uint32 `json:"可充电储能装置故障代码列表"`
	Electric_alarm_sum  uint8    `json:"驱动电机故障总数N2"`
	Electric_alarm_list []uint32 `json:"驱动电机故障代码列表"`
	Engine_alarm_sum    uint8    `json:"发动机故障总数N3"`
	Engine_alarm_list   []uint32 `json:"发动机故障代码列表"`
	Other_alarm_sum     uint8    `json:"其他故障总数N4"`
	Other_alarm_list    []uint32 `json:"其他故障代码列表"`
}

/*moduler voltage data*/
type PackvData struct {
	PackNo    uint8    `json:"可充电储能子系统号"`
	Packv     uint16   `json:"可充电储能装置电压"`
	Packc     uint16   `json:"可充电储能装置电流"`
	Monosum   uint16   `json:"单体电池总数"`
	SfSN      uint16   `json:"本帧起始电池序号"`
	Tfmonosum uint8    `json:"本帧单体电池总数"`
	Monov     []uint16 `json:"单体电池电压"`
}

type ModulevData struct {
	packsum uint8
	packvs  [DefCfg_Subsys_Num]PackvData
}

/*moduler temperature data*/
type PacktData struct {
	PackNo   uint8  `json:"可充电储能子系统号"`
	ProbeNum uint16 `json:"可充电储能温度探针个数"`
	//Probet   []uint8 `json:"可充电储能温度探针温度值"`
	Probet []int16 `json:"可充电储能温度探针温度值"`
}

type ModuletData struct {
	packsum uint8
	packts  [DefCfg_Subsys_Num]PacktData
}

// CheckCommandFlag check
// #define Command_Vehicle_Login		0x01
// #define Command_RealTime_Info		0x02
// #define Command_Reissue_Info			0x03
// #define Command_Vehicle_LoginOut		0x04
// #define Command_Platform_Login		0x05
// #define Command_Platform_LoginOut	0x06
// #define Command_Terminal_Beartbeat	0x07
// #define Command_Terminal_School		0x08
// #define Command_Platform_Query		0x80
// #define Command_Platform_Set			0x81
// #define Command_Platform_Contrl		0x82

// #define HSCommand_CHK				0x50
// #define HSCommand_CHK_GJ				0x51
// #define HSCommand_CHK_GJ_CFG_RP		0x52

func (p *GBT32960Protocol) CheckCommandFlag(cmd_flag byte) string {

	var msg_cmd string

	switch {
	case cmd_flag == 0x01:
		msg_cmd = "车辆登入"
	case cmd_flag == 0x02:
		msg_cmd = "实时信息上报"
	case cmd_flag == 0x03:
		msg_cmd = "补发信息上报"
	case cmd_flag == 0x04:
		msg_cmd = "车辆登出"
	case cmd_flag == 0x05:
		msg_cmd = "平台登入"
	case cmd_flag == 0x06:
		msg_cmd = "平台登出"
	case cmd_flag == 0x07:
		msg_cmd = "心跳"
	case cmd_flag == 0x08:
		msg_cmd = "终端数据预留"
	case cmd_flag >= 0x09 && cmd_flag <= 0x7F:
		msg_cmd = "上行数据系统预留"
	case cmd_flag >= 0x80 && cmd_flag <= 0x82:
		msg_cmd = "终端数据预留"
	case cmd_flag >= 0x83 && cmd_flag <= 0xBF:
		msg_cmd = "下行数据系统预留"
	case cmd_flag >= 0xC0 && cmd_flag <= 0xFE:
		msg_cmd = "平台交换自定义数据"
	default:
		msg_cmd = "---"
	}
	return msg_cmd
}

// CheckResponseFlag check
func (p *GBT32960Protocol) CheckResponseFlag(resp_flag byte) string {

	var msg_resp string

	switch {
	case resp_flag == 0x01:
		msg_resp = "成功"
	case resp_flag == 0x02:
		msg_resp = "错误"
	case resp_flag == 0x03:
		msg_resp = "VIN重复"
	case resp_flag == 0xFE:
		msg_resp = "命令"
	default:
		msg_resp = "---"
	}
	return msg_resp
}

// CheckDataType check
func (p *GBT32960Protocol) CheckDataType(data_type byte) string {
	var data_info string
	switch {
	case data_type == 0x01:
		data_info = "整车"
	case data_type == 0x02:
		data_info = "驱动电机"
	case data_type == 0x03:
		data_info = "燃料电池"
	case data_type == 0x04:
		data_info = "发动机"
	case data_type == 0x05:
		data_info = "车辆位置"
	case data_type == 0x06:
		data_info = "极值"
	case data_type == 0x07:
		data_info = "报警"
	// case data_type >= 0x08 && data_type <= 0x09:
	// 	data_info = "终端数据预留"
	case data_type == 0x08:
		//data_info = "可充电储能装置电压数据"
		data_info = "电压数据"
	case data_type == 0x09:
		//data_info = "可充电储能装置温度数据"
		data_info = "温度数据"
	case data_type >= 0x0A && data_type <= 0x2F:
		data_info = "平台交换协议自定义数据"
	case data_type >= 0x30 && data_type <= 0x7F:
		data_info = "预留"
	case data_type >= 0x80 && data_type <= 0xFE:
		// #define BMS_STAT_DATA_ID        0xA0
		// #define BMS_KEY_DATA_ID         0xA1
		// #define BMS_STAT_DATA_ID2       0xA2
		// #define BMS_KEY_DATA_ID2        0xA3
		// #define LMGM_GJ_DATA            0xB0
		data_info = "用户自定义"
	default:
		data_info = "---"
	}
	return data_info
}

// UnpackEVData converts a slice of bytes in logfmt format to metrics.
func (p *GBT32960Protocol) UnpackEVData(msg *GBT32960Message) (map[string]interface{}, error) {

	var result map[string]interface{}

	// 数据单元起始字节30，数据单元结束字节(msg.len-1)
	i := 30
	for {
		if i >= int(msg.len-1) {
			break
		}

		data_type := msg.body[i]
		//log.Printf("-> %d %x %v %d %x\n", i, data_type, p.CheckDataType(data_type), len(msg.body[i:]), msg.body[i:])
		i += 1

		switch {

		case data_type == 0x01: // "整车数据"

			data := &VcuData{
				Vehicle_status:  msg.body[i],
				Charging_status: msg.body[i+1],
				Operation_mode:  msg.body[i+2],
				Speed:           binary.BigEndian.Uint16(msg.body[i+3 : i+5]),
				Mileage:         binary.BigEndian.Uint32(msg.body[i+5 : i+9]),
				Voltage:         binary.BigEndian.Uint16(msg.body[i+9 : i+11]),
				Current:         binary.BigEndian.Uint16(msg.body[i+11 : i+13]),
				Soc:             msg.body[i+13],
				Dc_status:       msg.body[i+14],
				Stall:           msg.body[i+15],
				Resistance:      binary.BigEndian.Uint16(msg.body[i+16 : i+18]),
				Acc_pedal:       msg.body[i+18],
				Braking_status:  msg.body[i+19],
			}
			i += binary.Size(data)

			if j, err := json.Marshal(data); err == nil {
				// log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
				if err := json.Unmarshal([]byte(j), &result); err != nil {
					fmt.Println("JsonToMapDemo err: ", err)
					return nil, err
				}
			}

		case data_type == 0x02: //"驱动电机数据"

			num := int(msg.body[i])
			i += 1
			for index := 0; index < num; index++ {

				data := &MotorInfo{
					Motor_SN:     msg.body[i],
					Motor_status: msg.body[i+1],
					Motor_ctr:    msg.body[i+2],
					Motor_speed:  binary.BigEndian.Uint16(msg.body[i+3 : i+5]),
					Motororque:   binary.BigEndian.Uint16(msg.body[i+5 : i+7]),
					Motor:        msg.body[i+7],
					Motor_ctr_v:  binary.BigEndian.Uint16(msg.body[i+8 : i+10]),
					Motor_ctr_c:  binary.BigEndian.Uint16(msg.body[i+10 : i+12]),
				}
				i += binary.Size(data)

				if j, err := json.Marshal(data); err == nil {
					//log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
					if err := json.Unmarshal([]byte(j), &result); err != nil {
						fmt.Println("JsonToMapDemo err: ", err)
					}
				}
			}

		case data_type == 0x03: //"燃料电池数据"

			log.Printf("-> %04d %v\n", i, p.CheckDataType(data_type))

		case data_type == 0x04: //"发动机数据"

			log.Printf("m-> %04d %v\n", i, p.CheckDataType(data_type))

		case data_type == 0x05: //"车辆位置数据"

			data := &PositionData{
				Loc_status: msg.body[i],
				Longitude:  binary.BigEndian.Uint32(msg.body[i+1 : i+5]),
				Latitude:   binary.BigEndian.Uint32(msg.body[i+5 : i+9]),
			}
			i += binary.Size(data)

			// if j, err := json.Marshal(data); err == nil {
			// 	//log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
			// 	if err := json.Unmarshal([]byte(j), &result); err != nil {
			// 		fmt.Println("JsonToMapDemo err: ", err)
			// 	}
			// }

		case data_type == 0x06: //"极值数据"

			data := &ExtrmvData{
				Hv_subsys_num: msg.body[i],
				Hv_cell_num:   msg.body[i+1],
				Shv:           binary.BigEndian.Uint16(msg.body[i+2 : i+4]),
				Lv_subsys_num: msg.body[i+4],
				Lv_cell_num:   msg.body[i+5],
				Slv:           binary.BigEndian.Uint16(msg.body[i+6 : i+8]),
				Ht_subsys_num: msg.body[i+8],
				Ht_cell_num:   msg.body[i+9],
				Sht:           msg.body[i+10],
				Lt_subsys_num: msg.body[i+11],
				Lt_cell_num:   msg.body[i+12],
				Slt:           msg.body[i+13],
			}
			i += binary.Size(data)

			// if j, err := json.Marshal(data); err == nil {
			// 	//log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
			// 	if err := json.Unmarshal([]byte(j), &result); err != nil {
			// 		fmt.Println("JsonToMapDemo err: ", err)
			// 	}
			// }

		case data_type == 0x07: // "报警数据"

			data := &AlarmData{
				Highst_alart_level: msg.body[i],
				General_alarm_flag: binary.BigEndian.Uint32(msg.body[i+1 : i+5]),
			}
			i += 5

			data.Battery_alarm_sum = msg.body[i]
			if data.Battery_alarm_sum > 0 {
				data.Battery_alarm_list = make([]uint32, data.Battery_alarm_sum)
				for index := 0; index < int(data.Battery_alarm_sum); index++ {
					data.Battery_alarm_list[index] = binary.BigEndian.Uint32(msg.body[i+1+index*4 : i+1+index*4+4])
				}
				i += 1 + 4*int(data.Battery_alarm_sum)
			} else {
				i += 1
			}

			data.Electric_alarm_sum = msg.body[i]
			if data.Electric_alarm_sum > 0 {
				data.Electric_alarm_list = make([]uint32, data.Electric_alarm_sum)
				for index := 0; index < int(data.Electric_alarm_sum); index++ {
					data.Electric_alarm_list[index] = binary.BigEndian.Uint32(msg.body[i+1+index*4 : i+1+index*4+4])
				}
				i += 1 + 4*int(data.Electric_alarm_sum)
			} else {
				i += 1
			}

			data.Engine_alarm_sum = msg.body[i]
			if data.Engine_alarm_sum > 0 {
				data.Engine_alarm_list = make([]uint32, data.Engine_alarm_sum)
				for index := 0; index < int(data.Engine_alarm_sum); index++ {
					data.Engine_alarm_list[index] = binary.BigEndian.Uint32(msg.body[i+1+index*4 : i+1+index*4+4])
				}
				i += 1 + 4*int(data.Engine_alarm_sum)
			} else {
				i += 1
			}

			data.Other_alarm_sum = msg.body[i]
			if data.Other_alarm_sum > 0 {
				data.Other_alarm_list = make([]uint32, data.Other_alarm_sum)
				for index := 0; index < int(data.Other_alarm_sum); index++ {
					data.Other_alarm_list[index] = binary.BigEndian.Uint32(msg.body[i+1+index*4 : i+1+index*4+4])
				}
				i += 1 + 4*int(data.Other_alarm_sum)
			} else {
				i += 1
			}

			// if j, err := json.Marshal(data); err == nil {
			// 	log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
			// }

		case data_type == 0x08: //"可充电储能装置电压数据"

			num := int(msg.body[i])
			i++
			for index := 0; index < num; index++ {
				data := &PackvData{
					PackNo:    msg.body[i],
					Packv:     binary.BigEndian.Uint16(msg.body[i+1 : i+3]),
					Packc:     binary.BigEndian.Uint16(msg.body[i+3 : i+5]),
					Monosum:   binary.BigEndian.Uint16(msg.body[i+5 : i+7]),
					SfSN:      binary.BigEndian.Uint16(msg.body[i+7 : i+9]),
					Tfmonosum: msg.body[i+9],
				}
				i += 10

				if data.Tfmonosum > 0 {
					data.Monov = make([]uint16, data.Tfmonosum)
					for index2 := 0; index2 < int(data.Tfmonosum); index2++ {
						data.Monov[index2] = binary.BigEndian.Uint16(msg.body[i+index2*2 : i+index2*2+2])
					}
					i += 2 * int(data.Tfmonosum)
				} else {
					i += 0
				}

				// if j, err := json.Marshal(data); err == nil {
				// 	//log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
				// 	if err := json.Unmarshal([]byte(j), &result); err != nil {
				// 		fmt.Println("JsonToMapDemo err: ", err)
				// 	}
				// }
			}

		case data_type == 0x09: //"可充电储能装置温度数据"

			num := int(msg.body[i])
			i++
			for index := 0; index < num; index++ {
				data := &PacktData{
					PackNo:   msg.body[i],
					ProbeNum: binary.BigEndian.Uint16(msg.body[i+1 : i+3]),
				}
				i += 3

				if data.ProbeNum > 0 {
					data.Probet = make([]int16, data.ProbeNum)
					for index2 := 0; index2 < int(data.ProbeNum); index2++ {
						data.Probet[index2] = int16(msg.body[i+index2]) - 40
						// log.Printf("msg %d %d %v %v\n", i, index2, uint16(msg.body[i+index2]), data.Probet)
					}
					i += 1 * int(data.ProbeNum)
				} else {
					i += 0
				}

				// if j, err := json.Marshal(data); err == nil {
				// 	//log.Printf("-> %v %s\n", p.CheckDataType(data_type), j)
				// 	if err := json.Unmarshal([]byte(j), &result); err != nil {
				// 		fmt.Println("JsonToMapDemo err: ", err)
				// 	}
				// }
			}
		case data_type >= 0x0A && data_type <= 0x2F: //"平台交换协议自定义数据"
			log.Printf("-> %d %x %v\n", i, data_type, p.CheckDataType(data_type))

		case data_type >= 0x30 && data_type <= 0x7F: //"预留"
			log.Printf("-> %d %x %v\n", i, data_type, p.CheckDataType(data_type))

		case data_type >= 0x80 && data_type <= 0xFE: // "用户自定义"
			//log.Printf("-> %d %x %v\n", i, data_type, p.CheckDataType(data_type))
			if data_type == 0xA0 || data_type == 0xA1 || data_type == 0xA2 || data_type == 0xA3 {
				// TOOD: BMS关键数据
				customLen := binary.BigEndian.Uint16(msg.body[i : i+2])
				// log.Printf("-> TOOD: BMS关键数据: %s %x %d %x %d", msg.vin, data_type, len(msg.body[i:]), msg.body[i:], customLen)
				i += (2 + int(customLen))
			} else if data_type == 0xA4 {
				// TOOD: 换电矿卡数据
				customLen := binary.BigEndian.Uint16(msg.body[i : i+2])
				// log.Printf("-> TOOD: 换电矿卡数据: %s %x %d %x %d", msg.vin, data_type, len(msg.body[i:]), msg.body[i:], customLen)
				i += (2 + int(customLen))
			} else {
				log.Panicf("-> TOOD: 未解析自定义数据: %s %x %d %x", msg.vin, data_type, len(msg.body), msg.body)
			}
		default:
			// log.Printf("-> TOOD: unknown: %s %d %x", msg.vin, len(msg.body[i:]), msg.body[i:])
			log.Panicf("-> TOOD: 未知数据类型: %s %d %x", msg.vin, len(msg.body), msg.body)
			return nil, ErrGbt32960Protocol // TODO: 这些数据不能丢
		}
	}
	return result, nil
}

// UnpackEVLogin login
func (p *GBT32960Protocol) UnpackEVLogin(msg *GBT32960Message) (map[string]interface{}, error) {

	var result map[string]interface{}

	i := 30
	for {
		if i >= int(msg.len-1) {
			break
		}

		data := &EvData{
			Sn:            binary.BigEndian.Uint16(msg.body[i : i+2]),
			Iccid:         string(msg.body[i+2 : i+22]),
			Pack_count:    msg.body[i+22],
			Pack_code_len: msg.body[i+23],
			Pack_code:     string(msg.body[i+24 : msg.len-1]),
		}

		// if j, err := json.Marshal(data); err == nil {
		// 	log.Printf("-> login: %s %s\n", msg.vin, j)
		// }

		i += (2 + 20 + 1 + 1 + int(data.Pack_code_len)*int(data.Pack_count))
	}
	return result, nil
}

// UnpackEVLogout logout
func (p *GBT32960Protocol) UnpackEVLogout(msg *GBT32960Message) (map[string]interface{}, error) {

	var result map[string]interface{}

	i := 30
	for {
		if i >= int(msg.len-1) {
			break
		}

		// data := &EvLogout{
		// 	Sn: binary.BigEndian.Uint16(msg.body[i : i+2]),
		// }

		// if j, err := json.Marshal(data); err == nil {
		// 	log.Printf("-> logout: %s %s\n", msg.vin, j)
		// }

		i += 2
	}
	return result, nil
}
