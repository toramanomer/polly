import { Navigate, useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import {
	Alert,
	Box,
	Button,
	Card,
	CardContent,
	CircularProgress,
	Divider,
	LinearProgress,
	Sheet,
	Stack,
	Typography
} from '@mui/joy'

import { useState } from 'react'
import { CreatePoll } from './createPoll'

type PollOption = {
	id: string
	text: string
	position: number
	count: number
}

type Poll = {
	id: string
	question: string
	options: PollOption[]
	expiresAt: string
	createdAt: string
}

export const Home = () => {
	const query = useQuery({
		queryFn: async () => {
			console.log('queryFn')
			const response = await fetch('/api/auth/me')
			const body = await response.json()
			return body?.token
		},
		queryKey: ['token'],
		retry: false
	})
	console.log(query)
	const [showCreateForm, setShowCreateForm] = useState(false)
	const {
		data: polls,
		isLoading,
		error
	} = useQuery<Poll[]>({
		queryKey: ['userPolls'],
		queryFn: async () => {
			const response = await fetch('/api/polls')
			if (!response.ok) throw new Error('Failed to fetch polls')
			return response.json()
		},
		retry: false
	})

	const navigate = useNavigate()

	const handleLogout = async () => {
		await fetch('/api/auth/signout', {
			method: 'POST',
			credentials: 'include'
		})

		navigate('/signin', { replace: true })
	}

	if (query.isPending)
		return (
			<Box
				sx={{
					minHeight: '100vh',
					display: 'flex',
					justifyContent: 'center',
					alignItems: 'center'
				}}
			>
				<CircularProgress />
			</Box>
		)

	if (query.isError) return <Navigate to='/signin' replace />

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
					maxWidth: 800,
					p: 4,
					borderRadius: 'lg',
					boxShadow: 'lg'
				}}
			>
				<Stack spacing={3}>
					<Stack
						direction='row'
						justifyContent='space-between'
						alignItems='center'
					>
						<Typography level='h2'>My Polls</Typography>
						<Button
							onClick={() => setShowCreateForm(!showCreateForm)}
						>
							{showCreateForm ? 'Cancel' : 'Create New Poll'}
						</Button>
					</Stack>
					{showCreateForm ? (
						<CreatePoll
							onSuccess={() => setShowCreateForm(false)}
						/>
					) : (
						<Polls polls={polls} />
					)}
				</Stack>
			</Sheet>
		</Box>
	)
}

const PollOption = ({
	totalVoteCount,
	option
}: {
	totalVoteCount: number
	option: PollOption
}) => {
	const percentage =
		totalVoteCount > 0 ? (option.count / totalVoteCount) * 100 : 0

	return (
		<Box>
			<Stack
				direction='row'
				justifyContent='space-between'
				spacing={1}
				mb={1}
			>
				<Typography>{option.text}</Typography>
				<Typography>{option.count} votes</Typography>
			</Stack>
			<LinearProgress
				determinate
				value={percentage}
				sx={{ '--LinearProgress-thickness': '8px' }}
			/>
		</Box>
	)
}

const Polls = ({ polls }: { polls: Poll[] | undefined }) => {
	if (!polls?.length)
		return (
			<Alert color='neutral' variant='soft'>
				You haven't created any polls yet. Create your first poll to get
				started!
			</Alert>
		)

	return (
		<Stack spacing={3}>
			{polls.map(poll => {
				const isExpired = new Date(poll.expiresAt) < new Date()
				const createdDate = new Date(
					poll.createdAt
				).toLocaleDateString()
				const expiresDate = new Date(
					poll.expiresAt
				).toLocaleDateString()

				const totalVoteCount = poll.options.reduce(
					(totalVoteCount, option) => totalVoteCount + option.count,
					0
				)
				return (
					<Card key={poll.id} variant='outlined'>
						<CardContent>
							<Stack spacing={2}>
								<Stack
									direction='row'
									justifyContent='space-between'
									alignItems='center'
								>
									<Typography level='h4'>
										{poll.question}
									</Typography>
									<Typography level='body-sm'>
										Created: {createdDate}
									</Typography>
								</Stack>

								<Stack
									direction='row'
									spacing={1}
									alignItems='center'
								>
									<Typography level='body-sm'>
										Expires: {expiresDate}
									</Typography>
									{isExpired && (
										<Alert
											color='warning'
											variant='soft'
											size='sm'
										>
											Expired
										</Alert>
									)}
								</Stack>

								<Divider />

								<Stack spacing={2}>
									{poll.options.map(option => (
										<PollOption
											totalVoteCount={totalVoteCount}
											option={option}
										/>
									))}
								</Stack>

								<Stack
									direction='row'
									justifyContent='space-between'
									alignItems='center'
								>
									<Typography level='body-sm'>
										Total votes: {totalVoteCount}
									</Typography>
									<Button
										variant='outlined'
										size='sm'
										onClick={() => {}}
									>
										View Poll
									</Button>
								</Stack>
							</Stack>
						</CardContent>
					</Card>
				)
			})}
		</Stack>
	)
}
