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
      <div className="absolute start-4 pointer-events-none text-muted-foreground flex items-center">
        <Mail className="size-4.5" />
      </div>
      <input
        className="py-3.5 ps-11 pe-4 rounded-md border border-border focus:border-accent-brand focus:ring-2 focus:ring-accent-brand/20 bg-background text-foreground w-full outline-none transition-all shadow-xs text-sm sm:text-base placeholder:text-muted-foreground/60 disabled:opacity-60 disabled:cursor-not-allowed"
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
