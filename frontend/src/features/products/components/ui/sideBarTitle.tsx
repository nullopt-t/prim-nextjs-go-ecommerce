export default function SideBarTitle({ title }: { title: string }) {
  return (
    <span className="block font-medium text-foreground uppercase mb-2.5 text-xs tracking-wider">
      {title}
    </span>
  );
}
