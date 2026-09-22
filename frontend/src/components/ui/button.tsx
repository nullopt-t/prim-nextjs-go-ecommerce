"use client";

import React from "react";

interface CustomButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  text?: string;
  icon?: React.ReactNode;
  children?: React.ReactNode;
}

export function CustomButton({
  text,
  icon,
  children,
  onClick,
  className = "",
  ...props
}: CustomButtonProps) {
  return (
    <button
      className={`p-2.5 border border-border w-full h-full rounded-md font-medium text-txt-sm md:text-txt-md lg:text-txt-lg ${className}`}
      onClick={onClick}
      {...props}
    >
      {icon && <span className="inline-block mr-2">{icon}</span>}
      {text || children}
    </button>
  );
}
