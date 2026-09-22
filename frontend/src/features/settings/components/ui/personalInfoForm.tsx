"use client";

import { useState } from "react";
import { CustomInput } from "@/components/ui/input";
import { CustomButton } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { useTranslations } from "next-intl";

export default function PersonalInfoForm() {
  const [form, setForm] = useState({
    firstName: "Mohamed",
    lastName: "Mahmoud",
    email: "m.mahmoud.alsaid.official@gmail.com",
    phone: "+20123456789",
  });
  const t = useTranslations("settings");

  const infoFields = [
    {
      id: "first-name-1",
      labelTitle: "firstName",
      input: {
        type: "text",
        placeholder: "placeholders.firstName",
        value: form.firstName,
        isDisabled: false,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
          setForm((prev) => ({ ...prev, firstName: e.target.value }));
        },
      },
    },
    {
      id: "last-name-1",
      labelTitle: "lastName",
      input: {
        type: "text",
        placeholder: "placeholders.lastName",
        value: form.lastName,
        isDisabled: false,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
          setForm((prev) => ({ ...prev, lastName: e.target.value }));
        },
      },
    },
    {
      id: "email-1",
      labelTitle: "email",
      input: {
        type: "email",
        placeholder: "placeholders.email",
        value: form.email,
        isDisabled: true,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
          setForm((prev) => ({ ...prev, email: e.target.value }));
        },
      },
    },
    {
      id: "phone-1",
      labelTitle: "phone",
      input: {
        type: "text",
        placeholder: "placeholders.phone",
        value: form.phone,
        isDisabled: true,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
          setForm((prev) => ({ ...prev, phone: e.target.value }));
        },
      },
    },
  ];

  return (
    <form
      className="flex flex-col gap-4 max-w-xl"
      onSubmit={(e) => e.preventDefault()}
    >
      {infoFields.map((info) => (
        <label
          key={info.id}
          className="flex flex-col sm:flex-row sm:gap-5 sm:justify-between sm:items-center"
        >
          <div className="whitespace-nowrap mb-1 sm:mb-0">
            <Text text={t(`settings.${info.labelTitle}`)} />
          </div>
          <div className="sm:w-72 bg-input-background rounded-md border border-border">
            <CustomInput
              type={info.input.type}
              placeholder={t(`settings.${info.input.placeholder}`)}
              value={info.input.value}
              isDisabled={info.input.isDisabled}
              onChange={info.input.onChange}
            />
          </div>
        </label>
      ))}
      <div className="mt-4 sm:ml-auto w-full sm:w-40 bg-primary text-primary-foreground rounded-md hover:opacity-90">
        <CustomButton text={t("settings.saveChanges")} />
      </div>
    </form>
  );
}
