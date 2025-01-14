import React from 'react'
import PropTypes from 'prop-types'
import { Field } from 'formik'

const SelectType = ({ name, className, readOnly }) => {
  const list = () => {
    const validTypes = ['wait','waittemp', 'equipment', 'ato', 'temperature','directdoser', 'doser', 'timers', 'phprobes', 'subsystem', 'macro']
    return validTypes.map(item => {
      return (
        <option key={item} value={item}>
          {item}
        </option>
      )
    })
  }

  return (
    <Field
      name={name}
      component='select'
      className={`form-control ${className}`}
      disabled={readOnly}
    >
      <option value='' className='d-none'>-- Select Type --</option>
      {list()}
    </Field>
  )
}

SelectType.propTypes = {
  readOnly: PropTypes.bool,
  name: PropTypes.string,
  className: PropTypes.string
}

export default SelectType
