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
      className="p-2.5 pl-4 rounded-md border-2 border-border focus:border-accent-brand bg-input-background text-foreground w-full outline-none transition-colors"
      type={inputObj.type}
      value={inputObj.value}
      onChange={(e) => inputObj.setValue(e.target.value)}
      placeholder={t(inputObj.placeholder)}
    />
  );
}
