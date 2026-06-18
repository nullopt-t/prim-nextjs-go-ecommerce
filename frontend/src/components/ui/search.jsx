import { Search } from 'lucide-react';

export function SearchBar() {

    return (
        <div className='relative'>
            <Search className='absolute top-1/2 -translate-1/2 left-5 text-muted-foreground z-10' />
            <input type='text' className='w-full text-txt-sm md:text-txt-md lg:text-txt-lg text-black placeholder:text-muted-foreground bg-input-background rounded-lg p-2 pl-10 border-2 border-border' placeholder='Search...' />
        </div>
    )
}