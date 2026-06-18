import { SectionTitle, SectionGrid } from "@/features/home/components/ui"

export function Section() {
    return (
        <>
            <div className='mb-5'>
                <SectionTitle title='categories' />
            </div>
            <SectionGrid />
        </>
    )
}