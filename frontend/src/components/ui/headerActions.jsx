import { Heart, ShoppingCart, User, Moon } from 'lucide-react';
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from '@/context/theme'

export function HeaderActions() {
    const { theme, toggle } = useTheme();

    const icons = [
        {
            id: 1,
            icon: User
        },
        {
            id: 2,
            icon: Heart
        },
        {
            id: 3,
            icon: ShoppingCart
        },
    ];

    return (
        <>
            <p className='cursor-pointer font-medium'>EN/ع</p>
            {theme === 'light' ?
                <MdOutlineWbSunny onClick={toggle} className='size-6 cursor-pointer' />
                :
                <Moon onClick={toggle} className='size-6 cursor-pointer' />
            }
            {icons.map(value => (
                <value.icon key={value.id} className='size-6 cursor-pointer' />
            ))}
        </>
    )
}