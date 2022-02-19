package system

import (
	"time"

	"github.com/reef-pi/reef-pi/controller/modules/timer"
)

type Program struct {
	MainTankTemperature     string `json:"maintank_temperature"`
	ReservoirTemperature    string `json:"reservoir_temperature"`
	MainTankWaterLevel      string `json:"maintank_waterlevel"`
	ReservoirWaterLevel     string `json:"reservoir_waterlevel"`
	ConditionerWaterLevel   string `json:"conditioner_waterlevel"`
	Ph                      string `json:"ph"`
	Orp                     string `json:"orp"`
	MainTankHeaterRelay     string `json:"maintank_heater_relay"`
	ReservoirHeaterRelay    string `json:"reservoir_heater_relay"`
	Custom1Relay            string `json:"custom1_relay"`
	Custom2Relay            string `json:"custom2_relay"`
	WaterChangeDuration     string `json:"waterchange_duration"`
	WaterChangeSchedule     string `json:"waterchange_schedule"`
	ConditionerDuration     string `json:"conditioner_duration"`
	ReservoirFillUpDuration string `json:"reservoir_fill_up_duration"`
}
type MacroInitConfig struct {
	safeTemperatureRangeForReservoir []float64
	waterOutDuration                 uint
	waterOutSpeed                    uint8
	waitTempPeriod                   time.Duration
	mainTankATOTimeGuard             time.Duration
	reservoirATOTimeGuard            time.Duration
	conditionerDosingDuration        time.Duration
	conditionerDosingSpeed           uint8
}

type TimerInitConfig struct {
	waterChangeSchedules    []timer.Job
	reservoirFillUpSchedule timer.Job
}
