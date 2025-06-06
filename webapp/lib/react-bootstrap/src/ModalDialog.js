import classNames from 'classnames';
import React from 'react';
import PropTypes from 'prop-types';

import {
  bsClass,
  bsSizes,
  getClassSet,
  prefix,
  splitBsProps
} from './utils/bootstrapUtils';
import { Size } from './utils/StyleConfig';

const propTypes = {
  /**
   * A css class to apply to the Modal dialog DOM node.
   */
  dialogClassName: PropTypes.string,

  /**
   * A method to run for the mousedown event on the dialog.
   */
  handleDialogMouseDown: PropTypes.func
};

class ModalDialog extends React.Component {
  render() {
    const {
      dialogClassName,
      className,
      style,
      children,
      handleDialogMouseDown,
      ...props
    } = this.props;
    const [bsProps, elementProps] = splitBsProps(props);

    const bsClassName = prefix(bsProps);

    const modalStyle = { display: 'block', ...style };

    const dialogClasses = {
      ...getClassSet(bsProps),
      [bsClassName]: false,
      [prefix(bsProps, 'dialog')]: true
    };

    return (
      <div
        {...elementProps}
        tabIndex="-1"
        role="dialog"
        style={modalStyle}
        className={classNames(className, bsClassName)}
      >
        <div className={classNames(dialogClassName, dialogClasses)}>
          <div
            className={prefix(bsProps, 'content')}
            role="document"
            onMouseDown={handleDialogMouseDown}
          >
            {children}
          </div>
        </div>
      </div>
    );
  }
}

ModalDialog.propTypes = propTypes;

export default bsClass('modal', bsSizes([Size.LARGE, Size.SMALL], ModalDialog));
