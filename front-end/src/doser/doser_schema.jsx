import * as Yup from 'yup'

const DoserSchema = Yup.object().shape({
  name: Yup.string()
    .required('Name is required'),
  jack: Yup.string()
    .required('Jack is required'),
  pin: Yup.string()
    .required('Pin is required'),
  is_stepper: Yup.bool()
    .required('Is Stepper Motor? is required'),
  enable: Yup.bool()
    .required('Doser Status is required'),
  steps_per_revolution: Yup.number()
    .optional('Steps per revolution must be numeric'),
  duration: Yup.number()
    .required('Duration is required'),
  speed: Yup.number()
    .required('Speed is required'),
  month: Yup.string()
    .required('Month is required'),
  week: Yup.string()
    .required('Week is required'),
  day: Yup.string()
    .required('Day is required'),
  hour: Yup.string()
    .required('Hour is required'),
  minute: Yup.string()
    .required('Minute is required'),
  second: Yup.string()
    .required('Second is required')
})

export default DoserSchema
