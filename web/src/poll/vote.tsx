import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router'
import { Turnstile } from '@marsidev/react-turnstile'
import {
	Box,
	Button,
	Sheet,
	Stack,
	Typography,
	Radio,
	RadioGroup,
	FormControl,
	FormLabel,
	Alert,
	CircularProgress
} from '@mui/joy'
import { useMutation, useQuery } from '@tanstack/react-query'

type PollOption = {
	id: string
	text: string
	position: number
}

type Poll = {
	id: string
	question: string
	options: PollOption[]
	expiresAt: string
}

export const Vote = () => {
	const { pollId } = useParams<{ pollId: string }>()
	const [selectedOption, setSelectedOption] = useState<string>('')
	const [token, setToken] = useState<string>('')
	const {
		data: poll,
		isLoading,
		error
	} = useQuery<Poll>({
		queryKey: ['poll', pollId],
		queryFn: async () => {
			const response = await fetch(`/api/polls/${pollId}`)

			if (!response.ok) throw new Error('Failed to fetch poll')
			return response.json()
		},
		retry: false
	})

	const { mutate, isPending, isSuccess, isError } = useMutation({
		mutationFn: async (optionId: string) => {
			const response = await fetch(`/api/polls/${pollId}/vote`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'X-CF-Turnstile-Token': token
				},
				body: JSON.stringify({ optionId })
			})
			if (!response.ok) throw new Error('Failed to vote')
			return response.json()
		}
	})

	const handleVote = () => {
		if (!selectedOption || !token) return
		mutate(selectedOption)
	}

	if (isLoading) {
		return (
			<Box
				sx={{
					minHeight: '100vh',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center'
				}}
			>
				<CircularProgress />
			</Box>
		)
	}

	if (error || !poll) {
		return (
			<Box
				sx={{
					minHeight: '100vh',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center'
				}}
			>
				<Alert color='danger' variant='soft'>
					Failed to load poll. Please try again later.
				</Alert>
			</Box>
		)
	}

	const isExpired = new Date(poll.expiresAt) < new Date()

	return (
		<Box
			sx={{
				minHeight: '100vh',
				display: 'flex',
				flexDirection: 'column',
				alignItems: 'center',
				py: 4,
				px: 2
			}}
		>
			<Sheet
				sx={{
					width: '100%',
					maxWidth: 600,
					p: 4,
					borderRadius: 'lg',
					boxShadow: 'lg'
				}}
			>
				<Stack spacing={3}>
					<Typography level='h2' sx={{ textAlign: 'center' }}>
						{poll.question}
					</Typography>

					{isExpired ? (
						<Alert color='warning' variant='soft'>
							This poll has expired and is no longer accepting
							votes.
						</Alert>
					) : isSuccess ? (
						<Alert color='success' variant='soft'>
							Thank you for voting!
						</Alert>
					) : (
						<>
							<FormControl>
								<FormLabel>Select your answer</FormLabel>
								<RadioGroup
									value={selectedOption}
									onChange={e =>
										setSelectedOption(e.target.value)
									}
								>
									{poll.options
										.sort((a, b) => a.position - b.position)
										.map(option => (
											<Radio
												key={option.id}
												value={option.id}
												label={option.text}
												sx={{ mb: 1 }}
											/>
										))}
								</RadioGroup>
							</FormControl>

							<Button
								size='lg'
								onClick={handleVote}
								disabled={
									!selectedOption || isPending || !token
								}
								loading={isPending}
							>
								Submit Vote
							</Button>

							{isError && (
								<Alert color='danger' variant='soft'>
									Failed to submit vote. Please try again.
								</Alert>
							)}
							<Turnstile
								siteKey='0x4AAAAAABadXcRKZ2B7dbnQ'
								onSuccess={setToken}
								options={{
									refreshExpired: 'auto',
									size: 'flexible'
								}}
							/>
						</>
					)}
				</Stack>
			</Sheet>
		</Box>
	)
}
