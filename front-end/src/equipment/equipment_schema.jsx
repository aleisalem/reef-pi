import * as Yup from 'yup'

const EquipmentSchema = Yup.object().shape({
  name: Yup.string().required('Name is required'),
  is_remote: Yup.boolean(),
  on_cmd: Yup.string().when('is_remote', {
    is: true,
    then: Yup.string().required('On Cmd is required')
  }),
  off_cmd: Yup.string().when('is_remote', {
    is: true,
    then: Yup.string().required('Off Cmd is required')
  }),
  remote_type: Yup.string().when('is_remote', {
    is: true,
    then: Yup.string().required('Remote type is required')
  }),
  outlet: Yup.string().when('is_remote',{
    is: false,
    then: Yup.string().required('Outlet is required')
  })
})

export default EquipmentSchema
