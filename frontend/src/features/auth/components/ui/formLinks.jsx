
function FormLinks({ type }) {

    let isRemember = type === 'login' || type === 'register' ? true : false;

    let txt = type === 'login' ? 'Forgot password?' :
        type === 'register' || type === 'forget' ?
            'Already have an account?' : type === 'reset' ? 'Resend OTP' : '';

    return (
        <div className='flex justify-between items-center text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground'>
            {isRemember &&
                <label className='flex gap-2.5 items-center cursor-pointer hover:text-accent-foreground'>
                    <input className='size-4' type='checkbox' />
                    <p className=''>Remember me</p>
                </label>}

            <a className='capitalize cursor-pointer hover:text-accent-foreground font-medium ml-auto'>{txt}</a>
        </div>
    )
}

export default FormLinks;