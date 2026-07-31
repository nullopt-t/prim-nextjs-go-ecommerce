import { useTranslation } from "react-i18next";

function FormInput({ inputObj }) {
	const { t } = useTranslation("auth");
	return (
		<input
			className="p-2 pl-5 rounded-sm border-2 border-border focus:border-ring bg-input-background w-full"
			type={inputObj.type}
			value={inputObj.value}
			onChange={(e) => inputObj.setValue(e.target.value)}
			placeholder={t(inputObj.placeholder)}
		/>
	);
}

export default FormInput;
