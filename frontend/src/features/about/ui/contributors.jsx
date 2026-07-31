import { useTranslation } from "react-i18next";

export default function Contributors({ contObj }) {
	const { t } = useTranslation("about");

	return (
		<div className="flex flex-col items-center">
			<img
				src={contObj.img}
				alt={contObj.imgAlt}
				className="w-16 h-16 md:w-24 md:h-24 rounded-full object-cover object-center"
			/>
			<p className="max-w-16 truncate hover:whitespace-normal hover:overflow-visible hover:text-clip mt-2.5 font-medium text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
				{contObj.name}
			</p>
			<p className="text-sm text-muted-foreground">{t(contObj.role)}</p>
			<ul className="flex gap-2.5 mt-2.5">
				{contObj.links.map((link) => (
					<li
						key={link.id}
						className="text-lg text-muted-foreground hover:text-accent-brand"
					>
						<a href={link.link} target="_blank" className="">
							<link.icon />
						</a>
					</li>
				))}
			</ul>
		</div>
	);
}
