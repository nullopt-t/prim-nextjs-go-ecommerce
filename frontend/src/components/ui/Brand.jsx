import { useNavigate } from "react-router-dom";

export function Brand() {
	const navigate = useNavigate();

	return (
		<h1
			onClick={() => navigate("/")}
			className="cursor-pointer hover:scale-90 col-span-3 md:col-span-1 md:justify-self-start font-black text-title-sm md:text-title-md lg:text-title-lg text-center"
		>
			<span className="">PRI</span>
			<span className="text-accent-brand">M</span>
		</h1>
	);
}
