"use client";

import { useTranslations } from "next-intl";

interface FormInputProps {
  inputObj: {
    type: string;
    value: string;
    placeholder: string;
    setValue: (val: string) => void;
  };
}

export default function FormInput({ inputObj }: FormInputProps) {
  const t = useTranslations("auth");

  return (
    <input
      className="p-3 pl-4 rounded-xl border border-border focus:border-accent-brand focus:ring-2 focus:ring-accent-brand/20 bg-background text-foreground w-full outline-none transition-all shadow-xs text-sm"
      type={inputObj.type}
      value={inputObj.value}
      onChange={(e) => inputObj.setValue(e.target.value)}
      placeholder={t(inputObj.placeholder)}
    />
  );
}
