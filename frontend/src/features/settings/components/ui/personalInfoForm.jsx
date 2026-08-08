import { useState } from "react";
import { CustomInput, CustomButton, Text } from "@/components/ui";

export default function PersonalInfoForm() {
	const [form, setForm] = useState({
		firstName: "Mohamed",
		lastName: "Mahmoud",
		email: "m.mahmoud.alsaid.official@gmail.com",
		phone: "+20123456789",
	});

	const infoFields = [
		{
			id: "first-name-1",
			labelTitle: "First Name",
			input: {
				type: "text",
				placeholder: "Enter your first name",
				value: form.firstName,
				isDisabled: false,
				onChange: (e) => {
					setForm((prev) => ({ ...prev, firstName: e.target.value }));
				},
			},
		},
		{
			id: "last-name-1",
			labelTitle: "Last Name",
			input: {
				type: "text",
				placeholder: "Enter your last name",
				value: form.lastName,
				isDisabled: false,
				onChange: (e) => {
					setForm((prev) => ({ ...prev, lastName: e.target.value }));
				},
			},
		},
		{
			id: "email-1",
			labelTitle: "Email",
			input: {
				type: "email",
				placeholder: "Enter your email",
				value: form.email,
				isDisabled: true,
				onChange: (e) => {
					setForm((prev) => ({ ...prev, email: e.target.value }));
				},
			},
		},
		{
			id: "phone-1",
			labelTitle: "Phone",
			input: {
				type: "text",
				placeholder: "Enter your phone",
				value: form.phone,
				isDisabled: true,
				onChange: (e) => {
					setForm((prev) => ({ ...prev, phone: e.target.value }));
				},
			},
		},
	];

	return (
		<form
			className="flex flex-col gap-5"
			onSubmit={(e) => e.preventDefault()}
		>
			{infoFields.map((info) => (
				<label
					key={info.id}
					className="flex flex-col lg:flex-row lg:gap-5 lg:justify-between lg:items-center"
				>
					<div className="whitespace-nowrap mb-0.5">
						<Text text={info.labelTitle} />
					</div>
					<div className="lg:w-64 bg-input-background rounded-sm">
						<CustomInput
							type={info.input.type}
							placeholder={info.input.placeholder}
							value={info.input.value}
							isDisabled={info.input.isDisabled}
							onChange={info.input.onChange}
						/>
					</div>
				</label>
			))}
			<div className="ml-auto flex justify-center bg-primary text-primary-foreground hover:bg-accent hover:text-accent-foreground rounded-md">
				<CustomButton text="Save Changes" />
			</div>
		</form>
	);
}
