
function FormSubtitle({ type }) {

    const subTitle = type === 'login' ? 'Please enter your details to sign in to your account.'
        : type === 'register' ? 'Get started by creating your account in just a few steps.' :
            type === 'forget' ? `Enter your email address and we'll send you a link to reset your password.` :
                type === 'reset' ? 'We sent a 6-digit verification code to your email. Enter it below to proceed.' : '';

    return (
        <p className='text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg'>
            {subTitle}
        </p>
    )
}

export default FormSubtitle;