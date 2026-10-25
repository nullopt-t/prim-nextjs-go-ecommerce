import React from "react";

interface TextProps {
  text?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
}

export function Text({ text, children, className = "" }: TextProps) {
  return (
    <p
      className={`text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg ${className}`}
    >
      {text || children}
    </p>
  );
}
