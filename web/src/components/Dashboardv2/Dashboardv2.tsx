import React, { useState } from 'react';
import { FullscreenLayout } from '../layout';
import { App } from 'antd';

// 导入简化版工作流编辑器
import { SimpleReteEditor } from './SimpleReteEditor';
import { ConversationWorkflowAdapter } from './ConversationWorkflowAdapter';

const Dashboardv2: React.FC = () => {
  const { message } = App.useApp();
  const [conversationAdapter] = useState(() => new ConversationWorkflowAdapter());

  return (
    <FullscreenLayout>
      <div className="w-full h-full bg-gray-50 overflow-hidden relative">
          <div className="w-full h-full relative">
            <SimpleReteEditor
              workflowId="conversation-v1"
              adapter={conversationAdapter}
            />
          </div>
      </div>
    </FullscreenLayout>
  );
};

export default Dashboardv2;
