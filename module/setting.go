package module

import (
	"fmt"
	"strconv"
	"strings"
)

type SettingType string

const (
	SettingBool   SettingType = "bool"
	SettingFloat  SettingType = "float"
	SettingInt    SettingType = "int"
	SettingText   SettingType = "text"
	SettingSlider SettingType = "slider"
)

type Setting interface {
	Name() string
	Type() SettingType
	DisplayValue() string
	SetFromString(string) error
}

type BoolSetting struct {
	name  string
	Value bool
}

func NewBool(name string, def bool) *BoolSetting {
	return &BoolSetting{name: name, Value: def}
}

func (s *BoolSetting) Name() string         { return s.name }
func (s *BoolSetting) Type() SettingType    { return SettingBool }
func (s *BoolSetting) DisplayValue() string {
	if s.Value {
		return "§atrue"
	}
	return "§cfalse"
}
func (s *BoolSetting) SetFromString(v string) error {
	switch strings.ToLower(v) {
	case "true", "1", "on", "yes":
		s.Value = true
	case "false", "0", "off", "no":
		s.Value = false
	default:
		return fmt.Errorf("ожидается true/false")
	}
	return nil
}

type FloatSetting struct {
	name     string
	Value    float32
	Min, Max float32
}

func NewFloat(name string, def, min, max float32) *FloatSetting {
	return &FloatSetting{name: name, Value: def, Min: min, Max: max}
}

func (s *FloatSetting) Name() string         { return s.name }
func (s *FloatSetting) Type() SettingType    { return SettingFloat }
func (s *FloatSetting) DisplayValue() string { return fmt.Sprintf("§e%.2f", s.Value) }
func (s *FloatSetting) SetFromString(v string) error {
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return fmt.Errorf("ожидается число")
	}
	val := float32(f)
	if val < s.Min || val > s.Max {
		return fmt.Errorf("диапазон: %.2f – %.2f", s.Min, s.Max)
	}
	s.Value = val
	return nil
}

type IntSetting struct {
	name     string
	Value    int
	Min, Max int
}

func NewInt(name string, def, min, max int) *IntSetting {
	return &IntSetting{name: name, Value: def, Min: min, Max: max}
}

func (s *IntSetting) Name() string         { return s.name }
func (s *IntSetting) Type() SettingType    { return SettingInt }
func (s *IntSetting) DisplayValue() string { return fmt.Sprintf("§e%d", s.Value) }
func (s *IntSetting) SetFromString(v string) error {
	i, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("ожидается целое число")
	}
	if i < s.Min || i > s.Max {
		return fmt.Errorf("диапазон: %d – %d", s.Min, s.Max)
	}
	s.Value = i
	return nil
}

type TextSetting struct {
	name  string
	Value string
}

func NewText(name, def string) *TextSetting {
	return &TextSetting{name: name, Value: def}
}

func (s *TextSetting) Name() string         { return s.name }
func (s *TextSetting) Type() SettingType    { return SettingText }
func (s *TextSetting) DisplayValue() string { return "§f" + s.Value }
func (s *TextSetting) SetFromString(v string) error {
	s.Value = v
	return nil
}

type SliderSetting struct {
	name           string
	Value          float32
	Min, Max, Step float32
}

func NewSlider(name string, def, min, max, step float32) *SliderSetting {
	return &SliderSetting{name: name, Value: def, Min: min, Max: max, Step: step}
}

func (s *SliderSetting) Name() string      { return s.name }
func (s *SliderSetting) Type() SettingType { return SettingSlider }
func (s *SliderSetting) DisplayValue() string {
	if s.Step <= 0 {
		return fmt.Sprintf("§e%.2f", s.Value)
	}
	total := int((s.Max - s.Min) / s.Step)
	if total > 20 {
		total = 20
	}
	filled := int((s.Value - s.Min) / s.Step)
	if filled > total {
		filled = total
	}
	bar := ""
	for i := 0; i < total; i++ {
		if i < filled {
			bar += "§a█"
		} else {
			bar += "§8█"
		}
	}
	return fmt.Sprintf("%s §e%.2f", bar, s.Value)
}
func (s *SliderSetting) SetFromString(v string) error {
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return fmt.Errorf("ожидается число")
	}
	val := float32(f)
	if val < s.Min || val > s.Max {
		return fmt.Errorf("диапазон: %.2f – %.2f", s.Min, s.Max)
	}
	s.Value = val
	return nil
}
