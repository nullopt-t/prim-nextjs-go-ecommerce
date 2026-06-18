
export function Categories() {
    const categories = [
        {
            id: "cat_1",
            slug: "audio",
            name: {
                en: "Audio",
                ar: "الأجهزة الصوتية"
            }
        },
        {
            id: "cat_2",
            slug: "wearables",
            name: {
                en: "Wearables",
                ar: "الأجهزة القابلة للارتداء"
            }
        },
        {
            id: "cat_3",
            slug: "desk-setup",
            name: {
                en: "Desk Setup",
                ar: "مستلزمات المكتب"
            }
        },
        {
            id: "cat_4",
            slug: "accessories",
            name: {
                en: "Accessories",
                ar: "الإكسسوارات"
            }
        },
        {
            id: "cat_5",
            slug: "lighting",
            name: {
                en: "Lighting",
                ar: "وحدات الإضاءة"
            }
        }
    ];

    return (
        <p className='flex gap-2.5 '>
            {categories.map(value => (
                <span key={value.id} className='cursor-pointer'>{value.name.en}</span>
            ))}
        </p>
    )
} 