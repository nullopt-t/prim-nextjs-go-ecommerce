function FormInput({ inputObj }) {
	return (
		<input
			className="p-2 pl-5 rounded-sm border-2 border-border focus:border-ring bg-input-background w-full"
			key={inputObj.id}
			type={inputObj.type}
			value={inputObj.value}
			onChange={(e) => inputObj.setValue(e.target.value)}
			placeholder={inputObj.placeholder}
		/>
	);
}

export default FormInput;
