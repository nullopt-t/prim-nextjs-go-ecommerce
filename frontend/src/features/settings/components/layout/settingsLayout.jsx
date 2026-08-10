import { Title } from "@/components/ui/title";
import DangerZone from "@/features/settings/components/ui/danger";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";

export default function SettingsLayout() {
	return (
		<div className="">
			<Title
				title="Settings"
				subtitle="Manage your account preferences."
			/>
			<Title title="Personal Information" />
			<PersonalInfoForm />
			<Title title="Danger Zone" textColor="text-destructive" />
			<DangerZone />
		</div>
	);
}
