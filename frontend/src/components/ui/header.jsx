import { Search, Mic, Heart, ShoppingCart, User, Moon } from 'lucide-react';
import { Categories } from '@/components/ui/categories'

export function Header() {

    return (
        <div className=''>
            <div className='grid gap-5 grid-cols-3 p-5 pr-2.5 pl-2.5 border-b border-border-color'>
                <h1 className='font-black col-span-3 text-center'>
                    <span className=''>PRI</span>
                    <span className='text-orange-500'>M</span>
                </h1>
                <div className='col-span-3'>search</div>
                <div className='flex justify-end col-span-3 gap-2.5'>
                    <p className='cursor-pointer'>EN/ع</p>
                    <Moon />
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