"use client";

import { useTranslations } from "next-intl";
import { Mail } from "lucide-react";

interface FormInputProps {
  inputObj: {
    type: string;
    value: string;
    placeholder: string;
    setValue: (val: string) => void;
    disabled?: boolean;
  };
}

export default function FormInput({ inputObj }: FormInputProps) {
  const t = useTranslations("auth");

  return (
    <div className="relative flex items-center">
      <div className="absolute start-3 pointer-events-none text-muted-foreground flex items-center">
        <Mail className="size-4" />
      </div>
      <input
        className="w-full h-11 ps-9.5 pe-3.5 rounded-md border border-border bg-background text-foreground text-sm placeholder:text-muted-foreground/60 outline-none transition-colors focus:border-accent-brand focus:ring-1 focus:ring-accent-brand disabled:opacity-50 disabled:cursor-not-allowed"
        type={inputObj.type}
        value={inputObj.value}
        disabled={inputObj.disabled}
        onChange={(e) => inputObj.setValue(e.target.value)}
        placeholder={t(inputObj.placeholder)}
        autoComplete="email"
      />
    </div>
  );
}
