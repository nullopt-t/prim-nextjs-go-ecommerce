"use client";

import { Fragment, useState } from "react";
import FormButton from "@/features/auth/components/ui/formButton";
import FormTitle from "@/features/auth/components/ui/formTitle";
import FormSubtitle from "@/features/auth/components/ui/formSubtitle";
import FormInput from "@/features/auth/components/ui/formInput";
import { useTranslations } from "next-intl";

interface FormProps {
  formType: "login" | "verify" | "register";
  handleSubmit: (type: string, payload: { email: string; code: string }) => void;
}

export default function Form({ formType, handleSubmit }: FormProps) {
  const t = useTranslations("auth");

  const [inputs, setInputs] = useState({
    email: "",
    code: "",
  });

  const handleEmail = (emailValue: string) => {
    setInputs((prev) => ({ ...prev, email: emailValue }));
  };

  const handleCode = (codeValue: string) => {
    setInputs((prev) => ({ ...prev, code: codeValue }));
  };

  const inputTypes = [
    {
      id: "email-1",
      name: "sign.emailLabel",
      type: "email",
      placeholder: "sign.emailPlaceholder",
      value: inputs.email,
      setValue: handleEmail,
      isExist: formType === "login" || formType === "register",
    },
    {
      id: "code-1",
      name: "verify.codeLabel",
      type: "text",
      placeholder: "verify.codePlaceholder",
      value: inputs.code,
      setValue: handleCode,
      isExist: formType === "verify",
    },
  ];

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="mb-2">
        <FormTitle type={formType} />
      </div>
      <div className="mb-8">
        <FormSubtitle type={formType} />
      </div>

      <form className="flex flex-col gap-5">
        {inputTypes.map((val) => (
          <Fragment key={val.id}>
            {val.isExist && (
              <label className="block">
                <span className="block capitalize font-medium text-foreground mb-2 text-txt-sm md:text-txt-md">
                  {t(val.name)}
                </span>
                <FormInput inputObj={val} />
              </label>
            )}
          </Fragment>
        ))}

        <div className="mt-2">
          <FormButton
            type={formType}
            payload={inputs}
            handle={handleSubmit}
          />
        </div>
      </form>
    </div>
  );
}
