import React from "react";

interface CustomInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  isDisabled?: boolean;
}

export function CustomInput({
  type = "text",
  value,
  isDisabled,
  placeholder,
  onChange,
  className = "",
  ...props
}: CustomInputProps) {
  return (
    <input
      type={type}
      placeholder={placeholder}
      onChange={onChange}
      value={value}
      disabled={isDisabled}
      className={`p-2.5 truncate w-full text-txt-sm md:text-txt-md lg:text-txt-lg placeholder:text-muted-foreground disabled:text-muted-foreground bg-transparent text-foreground ${className}`}
      {...props}
    />
  );
}
