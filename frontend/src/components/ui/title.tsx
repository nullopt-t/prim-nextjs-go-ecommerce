import React from "react";

interface TitleProps {
  title: string;
  subtitle?: string;
  textColor?: string;
  className?: string;
}

export function Title({ title, subtitle, textColor, className = "" }: TitleProps) {
  return (
    <div className={`flex flex-col mb-5 ${className}`}>
      <span
        className={`${
          textColor || "text-foreground"
        } capitalize font-medium text-title-sm md:text-title-md lg:text-title-lg`}
      >
        {title}
      </span>
      {subtitle && (
        <span className="text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground">
          {subtitle}
        </span>
      )}
    </div>
  );
}
