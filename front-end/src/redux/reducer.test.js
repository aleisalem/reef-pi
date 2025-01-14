import { rootReducer } from './reducer'
import { configureStore } from './store'
describe('Redux Reducer', () => {
  it('Store', () => {
    configureStore()
  })
  it('reducer', () => {
    function getPayload () {
      return {
        foo: 'bar',
        id: 1,
        data: 'foobar:data',
        usage: 'foobar:usage',
        readings: 'foobar:readings'
      }
    }
    function getState () {
      return {
        camera: {},
        ato_usage: [],
        macro_usage: {},
        tc_usage: {},
        ph_readings: {}
      }
    }
    console.log = jest.fn()
    let result
    result = rootReducer(getState(), { type: 'ERRORS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), errors: getPayload() })
    result = rootReducer(getState(), { type: 'INFO_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), info: getPayload() })
    result = rootReducer(getState(), { type: 'TELEMETRY_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), telemetry: getPayload() })
    result = rootReducer(getState(), { type: 'TIMERS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), timers: getPayload() })
    result = rootReducer(getState(), { type: 'ATOS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), atos: getPayload() })
    result = rootReducer(getState(), { type: 'ATO_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), config: getPayload() })
    result = rootReducer(getState(), { type: 'ATO_USAGE_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), ato_usage: [undefined, 'foobar:data'] })
    result = rootReducer(getState(), { type: 'MACROS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), macros: getPayload() })
    result = rootReducer(getState(), { type: 'MACROS_SCHEDULED_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), macros_scheduled: getPayload() })
    result = rootReducer(getState(), { type: 'MACRO_USAGE_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), macro_usage: { 1: 'foobar:data' } })
    result = rootReducer(getState(), { type: 'TCS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), tcs: getPayload() })
    result = rootReducer(getState(), { type: 'TC_SENSORS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), tc_sensors: getPayload() })
    result = rootReducer(getState(), { type: 'TC_USAGE_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), tc_usage: { 1: 'foobar:usage' } })
    result = rootReducer(getState(), { type: 'LIGHTS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), lights: getPayload() })
    result = rootReducer(getState(), { type: 'DASHBOARD_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), dashboard: getPayload() })
    result = rootReducer(getState(), { type: 'PH_PROBES_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), phprobes: getPayload() })
    result = rootReducer(getState(), { type: 'PH_PROBE_READINGS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), ph_readings: { 1: 'foobar:readings' } })
    result = rootReducer(getState(), { type: 'CAPABILITIES_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), capabilities: getPayload() })
    result = rootReducer(getState(), { type: 'SETTINGS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), settings: getPayload() })
    result = rootReducer(getState(), { type: 'JACKS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), jacks: getPayload() })
    result = rootReducer(getState(), { type: 'INLETS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), inlets: getPayload() })
    result = rootReducer(getState(), { type: 'OUTLETS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), outlets: getPayload() })
    result = rootReducer(getState(), { type: 'EQUIPMENTS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), equipment: getPayload() })
    result = rootReducer(getState(), { type: 'HEALTH_STATS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), health_stats: getPayload() })
    result = rootReducer(getState(), { type: 'DISPLAY_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), display: getPayload() })
    result = rootReducer(getState(), { type: 'IMAGES_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), camera: { images: getPayload() } })
    result = rootReducer(getState(), { type: 'LATEST_IMAGE_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), camera: { latest: getPayload() } })
    result = rootReducer(getState(), { type: 'CAMERA_CONFIG_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), camera: { config: getPayload() } })
    result = rootReducer(getState(), { type: 'DOSING_PUMPS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), dosers: getPayload() })
    result = rootReducer(getState(), { type: 'ERRORS_LOADED', payload: getPayload() })
    expect(result).toEqual({ ...getState(), errors: getPayload() })
    result = rootReducer(getState(), { type: 'CREDS_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'EQUIPMENT_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'RELOADED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'REBOOTED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'POWER_OFFED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DASHBOARD_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'SETTINGS_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DISPLAY_SWITCHED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'BRIGHTNESS_SET' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DOSING_PUMP_CREATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DOSING_PUMP_DELETED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DOSING_PUMP_CALIBRATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'ATO_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'ATO_DELETED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'MACRO_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'MACRO_DELETED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'DOSING_PUMP_SCHEDULE_UPDATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'TIMER_CREATED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'TIMER_DELETED' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'MACRO_RUN' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'API_FAILURE' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'TELEMETRY_TEST_MESSAGE_SENT' })
    expect(result).toEqual(getState())
    result = rootReducer(getState(), { type: 'foo' })
    expect(result).toEqual(getState())
    expect(console.log.mock.calls.length).toBe(1)
  })
})
