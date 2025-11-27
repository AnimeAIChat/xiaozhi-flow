import React from 'react';

interface FullscreenLayoutProps {
  children: React.ReactNode;
}

const FullscreenLayout: React.FC<FullscreenLayoutProps> = ({
  children,
}) => {
  return (
    <div className="w-screen h-screen overflow-hidden">
      {children}
    </div>
  );
};

export default FullscreenLayout;