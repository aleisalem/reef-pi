import EditEquipment from './edit_equipment'
import EquipmentSchema from './equipment_schema'
import { withFormik } from 'formik'

const EditEquipmentForm = withFormik({
  displayName: 'EquipmentForm',
  mapPropsToValues: props => ({
    name: (props.equipment && props.equipment.name) || '',
    outlet: (props.equipment && props.equipment.outlet) || '',
    id: (props.equipment && props.equipment.id) || '',
    on: (props.equipment && props.equipment.on) || false,
    is_remote: (props.equipment && props.equipment.is_remote) || false,
    remote_type: (props.equipment && props.equipment.remote_type) || '',
    on_cmd: (props.equipment && props.equipment.on_cmd) || '',
    off_cmd: (props.equipment && props.equipment.off_cmd) || '',
    outlets: props.outlets,
    remove: props.remove
  }),
  validationSchema: EquipmentSchema,
  handleSubmit: (values, { props }) => {
    props.onSubmit(values)
  }
})(EditEquipment)

export default EditEquipmentForm
