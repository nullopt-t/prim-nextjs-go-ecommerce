import { Fragment, useState } from "react";
import FormButton from "@/features/auth/components/ui/formButton";
import FormTitle from "@/features/auth/components/ui/formTitle";
import FormSubtitle from "@/features/auth/components/ui/formSubtitle";
import FormInput from "@/features/auth/components/ui/formInput";
import { useTranslation } from "react-i18next";

function Form({ formType, handleSubmit }) {
	const { t } = useTranslation("auth");

	const [inputs, setInputs] = useState({
		email: "",
		code: "",
	});

	const handleEmail = (emailValue) => {
		setInputs((prev) => {
			return { ...prev, email: emailValue };
		});
	};

	const handleCode = (codeValue) => {
		setInputs((prev) => {
			return { ...prev, code: codeValue };
		});
	};

	const inputTypes = [
		{
			id: "email-1",
			name: "sign.emailLabel",
			type: "email",
			placeholder: "sign.emailPlaceholder",
			value: inputs.email,
			setValue: handleEmail,
			isExist:
				formType === "login" || formType === "register" ? true : false,
		},
		{
			id: "code-1",
			name: "verify.codeLabel",
			type: "text",
			placeholder: "verify.codePlaceholder",
			value: inputs.code,
			setValue: handleCode,
			isExist: formType === "verify" ? true : false,
		},
	];

	return (
		<div className="p-5">
			<div className="mb-2.5">
				<FormTitle type={formType} />
			</div>
			<div className="mb-10">
				<FormSubtitle type={formType} />
			</div>

			<form className="flex flex-col gap-7.5">
				{inputTypes.map((value) => (
					<Fragment key={value.id}>
						{value.isExist && (
							<label className="">
								<p
									className={`${formType === "register" ? 'after:content-["*"] after:ml-0.5 after:text-red-500' : ""} capitalize font-medium text-foreground mb-2.5 text-txt-sm md:text-txt-md lg:text-txt-lg`}
								>
									{t(value.name)}
								</p>
								<FormInput inputObj={value} />
							</label>
						)}
					</Fragment>
				))}

				<FormButton
					type={formType}
					payload={inputs}
					handle={handleSubmit}
				/>
			</form>
		</div>
	);
}

export default Form;
