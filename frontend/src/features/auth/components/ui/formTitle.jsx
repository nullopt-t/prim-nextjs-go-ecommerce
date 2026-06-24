import { House, Moon } from 'lucide-react'
import { MdOutlineWbSunny } from "react-icons/md"
import { useTheme } from '@/context/theme'

function FormTitle({ type }) {
    const { theme, toggle } = useTheme();

    const title = type === 'login' ? 'welcome back.'
        : type === 'register' ? 'Create an Account' :
            type === 'forget' ? 'Forgot Password?' :
                type === 'reset' ? 'Verify OTP Code' : '';

    return (
        <p className='flex justify-between items-center capitalize text-foreground font-medium text-title-sm md:text-title-md lg:text-title-lg'>
            <span className=''>{title}</span>

            <span className='flex items-center gap-2.5 text-muted-foreground'>
                <span className='hover:scale-85'>
                    <House className='size-6 cursor-pointer' />
                </span>
                <span className='hover:scale-85'>{theme === 'light' ?
                    <MdOutlineWbSunny onClick={toggle} className='size-6 cursor-pointer' />
                    :
                    <Moon onClick={toggle} className='size-6 cursor-pointer' />
                }</span>
            </span>
        </p>
    )
}

export default FormTitle;