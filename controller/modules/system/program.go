package system

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
