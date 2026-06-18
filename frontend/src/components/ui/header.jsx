import { Search, Mic, Heart, ShoppingCart, User, Moon } from 'lucide-react';
import { Categories } from '@/components/ui/categories'
import { useTheme } from '@/context/theme'

export function Header() {
    const { theme, toggle } = useTheme();

    return (
        <div className='text-foreground'>
            <div className='grid gap-5 grid-cols-3 p-5 pr-2.5 pl-2.5 border-b border-border-color'>
                <h1 className='font-black text-2xl md:text-3xl lg:text-4xl col-span-3 text-center'>
                    <span className=''>PRI</span>
                    <span className='text-orange-500'>M</span>
                </h1>
                <input type='text' className='text-black placeholder:text-muted-foreground col-span-3 bg-input-background rounded-lg p-2 pl-5 border-2 border-border' placeholder='Search...' />
                <div className='flex justify-end col-span-3 gap-5 text-foreground'>
                    <p className='cursor-pointer font-medium'>EN/ع</p>
                    <Moon onClick={toggle} />
                    <User />
                    <Heart />
                    <ShoppingCart />
                </div>
            </div>
            <div className='p-5 pr-2.5 pl-2.5'>
                <Categories />
            </div>
        </div>
    )
}