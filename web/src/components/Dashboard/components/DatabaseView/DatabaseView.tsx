import React from 'react';
import { DatabaseTableNodes } from '../../../../components/DatabaseTableNodes';
import { DatabaseViewProps } from '../../types';

const DatabaseView: React.FC<DatabaseViewProps> = ({
  schema,
  onTableSelect,
  onDoubleClick,
}) => {
  return (
    <div
      className="w-full h-full cursor-pointer"
      onDoubleClick={onDoubleClick}
      title="双击进入配置编辑器"
    >
      <DatabaseTableNodes
        schema={schema}
        onTableSelect={onTableSelect}
      />
    </div>
  );
};

export default DatabaseView;