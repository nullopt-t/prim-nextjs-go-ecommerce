import CustomInput from "@/components/ui/input";
import CustomButton from "@/components/ui/button";

function Copoun() {
	return (
		<div className="flex gap-10">
			<div className="flex-1 p-2.5 pl-5 border border-border rounded-md">
				<CustomInput
					type={"text"}
					placeholder={"Enter your copoun"}
					handle={() => null}
				/>
			</div>
			<div className="w-24 md:w-32">
				<CustomButton text={"Apply"} />
			</div>
		</div>
	);
}

export default Copoun;
