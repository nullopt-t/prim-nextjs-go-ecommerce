import { useNavigate } from "react-router-dom";
import { House, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useTranslation } from "react-i18next";

function FormTitle({ type }) {
	const { theme, toggle } = useTheme();
	const navigate = useNavigate();

	const { t } = useTranslation("auth");
	const title =
		type === "login"
			? t("sign.title")
			: type === "verify"
				? t("verify.title")
				: "";

	return (
		<p className="flex justify-between items-center capitalize text-foreground font-medium text-title-sm md:text-title-md lg:text-title-lg">
			<span className="">{title}</span>

			<span className="flex items-center gap-2.5 text-muted-foreground">
				<span className="hover:scale-85 hover:text-accent-foreground">
					<House
						onClick={() => navigate("/home")}
						className="size-6 cursor-pointer"
					/>
				</span>
				<span className="hover:scale-85 hover:text-accent-foreground">
					{theme === "light" ? (
						<MdOutlineWbSunny
							onClick={toggle}
							className="size-6 cursor-pointer"
						/>
					) : (
						<Moon
							onClick={toggle}
							className="size-6 cursor-pointer"
						/>
					)}
				</span>
			</span>
		</p>
	);
}

export default FormTitle;
