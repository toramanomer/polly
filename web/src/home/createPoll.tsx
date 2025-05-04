import { useForm, useFieldArray } from 'react-hook-form'
import { useMutation } from '@tanstack/react-query'
import {
	Button,
	Divider,
	FormControl,
	FormHelperText,
	FormLabel,
	IconButton,
	Input,
	Stack
} from '@mui/joy'

import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'

type CreatePollForm = {
	question: string
	options: { value: string }[]
	expiresAt: string
}

type CreatePollProps = {
	onSuccess: () => void
}

export const CreatePoll = ({ onSuccess }: CreatePollProps) => {
	const { register, handleSubmit, formState, control } =
		useForm<CreatePollForm>({
			defaultValues: {
				question: '',
				options: [{ value: '' }, { value: '' }],
				expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000)
					.toISOString()
					.slice(0, 16) // Default to 24 hours from now
			}
		})

	const { fields, append, remove } = useFieldArray({
		control,
		name: 'options'
	})

	const { mutate } = useMutation({
		mutationFn: async (data: CreatePollForm) => {
			const response = await fetch('/api/polls', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					question: data.question,
					options: data.options.map(opt => opt.value),
					expiresAt: new Date(data.expiresAt).toISOString()
				})
			})

			if (response.ok) return await response.json()
			throw await response.json()
		},
		onSuccess: () => {
			onSuccess()
		},
		onError: () => {
			console.log('error')
		}
	})

	const onSubmit = (data: CreatePollForm) => {
		mutate(data)
	}

	return (
		<form onSubmit={handleSubmit(onSubmit)}>
			<Stack spacing={3}>
				<FormControl error={!!formState.errors.question}>
					<FormLabel>Poll Question</FormLabel>
					<Input
						{...register('question', {
							required: 'Question is required'
						})}
						placeholder="What's your question?"
						size='lg'
					/>
					<FormHelperText>
						{formState.errors.question?.message}
					</FormHelperText>
				</FormControl>

				<Divider>Options</Divider>

				{fields.map((field, index) => (
					<FormControl key={field.id}>
						<Stack direction='row' spacing={1}>
							<Input
								{...register(
									`options.${index}.value` as const,
									{
										required: 'Option is required'
									}
								)}
								placeholder={`Option ${index + 1}`}
								size='lg'
								sx={{ flex: 1 }}
							/>
							{fields.length > 2 && (
								<IconButton
									variant='soft'
									color='danger'
									onClick={() => remove(index)}
								>
									<DeleteIcon />
								</IconButton>
							)}
						</Stack>
					</FormControl>
				))}

				{fields.length < 6 && (
					<Button
						variant='outlined'
						startDecorator={<AddIcon />}
						onClick={() => append({ value: '' })}
					>
						Add Option
					</Button>
				)}

				<Divider>Expiration</Divider>

				<FormControl>
					<FormLabel>When should this poll expire?</FormLabel>
					<Input
						{...register('expiresAt', {
							required: 'Expiration time is required'
						})}
						type='datetime-local'
						size='lg'
					/>
					<FormHelperText>
						The poll will automatically close at this time
					</FormHelperText>
				</FormControl>

				<Button type='submit' size='lg' sx={{ mt: 2 }}>
					Create Poll
				</Button>
			</Stack>
		</form>
	)
}
