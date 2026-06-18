import { FaInstagram, FaSquareFacebook, FaSquareXTwitter, FaTiktok } from "react-icons/fa6"

const currentDate = new Date().getFullYear();

export function Footer() {

    const sections = [
        {
            id: 'shop-pp',
            title: 'shop',
            content: [
                {
                    id: "cat_1",
                    slug: "audio",
                    name: "Audio"
                },
                {
                    id: "cat_2",
                    slug: "wearables",
                    name: "Wearables"
                },
                {
                    id: "cat_3",
                    slug: "desk-setup",
                    name: "Desk Setup",
                },
            ]
        },
        {
            id: 'support-pp',
            title: 'support',
            content: [
                {
                    id: "supp-1",
                    name: 'contact us'
                },
                {
                    id: "supp_2",
                    name: 'FAQs'
                },
                {
                    id: "supp_3",
                    name: 'shipping'
                },
                {
                    id: "supp_4",
                    name: 'returns'
                },
                {
                    id: "supp_5",
                    name: 'track order'
                },
                {
                    id: "supp_6",
                    name: 'size guide'
                },
            ]
        },
        {
            id: 'company-pp',
            title: 'company',
            content: [
                {
                    id: "comp-1",
                    name: 'about us'
                },
                {
                    id: "comp-2",
                    name: 'careers'
                },
                {
                    id: "comp-3",
                    name: 'press'
                },
                {
                    id: "comp-4",
                    name: 'Sustainability'
                },
                {
                    id: "comp-5",
                    name: 'blog'
                }
            ]
        },
    ];

    const social = [
        {
            id: 'insta-1',
            icon: FaInstagram
        },
        {
            id: 'x-1',
            icon: FaSquareXTwitter
        },
        {
            id: 'tik-1',
            icon: FaTiktok
        },
        {
            id: 'facebook-1',
            icon: FaSquareFacebook
        },
    ];

    return (
        <div className='bg-[#111111] p-2.5 pt-5 pb-5'>
            <div className='mb-5 flex flex-col gap-10 md:flex-row text-white text-txt-sm md:text-txt-md lg:text-txt-lg'>
                <div className='flex-1'>
                    <p className='font-medium text-title-sm md:text-title-md'>
                        <span className='text-white'>PRI</span>
                        <span className='text-orange-500'>M</span>
                    </p>
                    <p className='text-muted-foreground'>
                        Your one-stop destination for everything you need. Quality products, fast delivery, and exceptional service since 2020.
                    </p>
                </div>
                {sections.map(value => (
                    <div key={value.id} className='flex-1'>
                        <p className='mb-5 text-txt-lg font-medium capitalize'>{value.title}</p>
                        <p className='flex flex-col gap-2.5'>
                            {value.content.map(content => (
                                <span key={content.id} className='capitalize text-muted-foreground cursor-pointer hover:text-accent'>{content.name}</span>
                            ))}
                        </p>
                    </div>
                ))}
            </div>
            <hr className='text-muted-foreground' />
            <div className='flex flex-col gap-5 md:flex-row md:justify-between text-muted-foreground pt-5'>
                <p className=''>
                    &copy; {currentDate} PRIM. All rights reserved.
                </p>
                <div className='flex gap-2.5'>
                    {social.map(value => (
                        <value.icon key={value.id} className='size-6 hover:text-accent' />
                    ))}
                </div>
            </div>
        </div>
    )
}