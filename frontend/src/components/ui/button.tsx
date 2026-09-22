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
      className={`p-2.5 rounded-xl font-medium text-txt-sm md:text-txt-md transition-all cursor-pointer ${
        className.includes("bg-") ? "" : "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
      } ${className}`}
      onClick={onClick}
      {...props}
    >
      {icon && <span className="inline-block mr-2">{icon}</span>}
      {text || children}
    </button>
  );
}
