package calib

import "gopkg.in/yaml.v3"

// Calib is the parsed joyCalib.yaml contents.
type Calib struct {
	Devices []CalibDevice
}

// CalibDevice is one calibrated device.
type CalibDevice struct {
	Name         string
	InstanceGUID string
	ProductGUID  string
	Axes         []CalibAxis
}

// CalibAxis is one calibrated axis.
type CalibAxis struct {
	Axis   int
	Name   string
	Min    int
	Center int
	Max    int
}

type calibFile struct {
	CalibrationInfo struct {
		DeviceList []struct {
			DeviceName   string `yaml:"DeviceName"`
			InstanceGUID string `yaml:"InstanceGUID"`
			ProductGUID  string `yaml:"ProductGUID"`
			AxisList     []struct {
				Axis        int    `yaml:"Axis"`
				AxisName    string `yaml:"AxisName"`
				CalibMin    int    `yaml:"CalibMin"`
				CalibCenter int    `yaml:"CalibCenter"`
				CalibMax    int    `yaml:"CalibMax"`
			} `yaml:"AxisList"`
		} `yaml:"DeviceList"`
	} `yaml:"CalibrationInfo"`
}

// Parse unmarshals a joyCalib.yaml document.
func Parse(data []byte) (*Calib, error) {
	var raw calibFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	c := &Calib{}
	for _, d := range raw.CalibrationInfo.DeviceList {
		cd := CalibDevice{
			Name:         d.DeviceName,
			InstanceGUID: d.InstanceGUID,
			ProductGUID:  d.ProductGUID,
		}
		for _, a := range d.AxisList {
			cd.Axes = append(cd.Axes, CalibAxis{
				Axis:   a.Axis,
				Name:   a.AxisName,
				Min:    a.CalibMin,
				Center: a.CalibCenter,
				Max:    a.CalibMax,
			})
		}
		c.Devices = append(c.Devices, cd)
	}
	return c, nil
}

// Validate reports whether data is a parseable joyCalib.yaml document.
// An empty DeviceList is valid.
func Validate(data []byte) error {
	_, err := Parse(data)
	return err
}
