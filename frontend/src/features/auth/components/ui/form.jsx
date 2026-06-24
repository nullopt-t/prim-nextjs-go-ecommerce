import FormButton from '@/features/auth/components/ui/formButton'
import FormTitle from '@/features/auth/components/ui/formTitle'
import FormSubtitle from '@/features/auth/components/ui/formSubtitle'
import FormInput from '@/features/auth/components/ui/formInput'
import FormLinks from '@/features/auth/components/ui/formLinks'

function Form({ formType }) {

    const inputTypes = [
        {
            id: 1,
            name: 'email',
            type: 'email',
            placeholder: 'Enter your email',
            isExist: formType === 'login' || formType === 'register' || formType === 'forget' ? true : false
        },
        {
            id: 2,
            name: 'password',
            type: 'password',
            placeholder: 'Enter your password',
            isExist: formType === 'login' || formType === 'register' ? true : false
        },
        {
            id: 3,
            name: 'confirm password',
            type: 'password',
            placeholder: 'Confirm your password',
            isExist: formType === 'register' ? true : false
        }
    ];

    return (
        <div className='p-5'>
            <div className='mb-2.5'>
                <FormTitle type={formType} />
            </div>
            <div className='mb-10'>
                <FormSubtitle type={formType} />
            </div>

            <form className='flex flex-col gap-7.5'>
                {inputTypes.map(value => (
                    <label key={value.id} className=''>
                        {value.isExist &&
                            <>
                                <p className='capitalize text-foreground mb-2.5 text-txt-sm md:text-txt-md lg:text-txt-lg'>{value.name}</p>
                                <FormInput inputObj={value} />
                            </>
                        }
                    </label>
                ))}

                <FormLinks type={formType} />
                <FormButton type={formType} />
            </form>

        </div>
    )
}

export default Form;