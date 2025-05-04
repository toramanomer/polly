import { useState } from 'react'
import { Navigate, useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
	Alert,
	Box,
	Button,
	Card,
	CardContent,
	CircularProgress,
	Divider,
	LinearProgress,
	Modal,
	ModalClose,
	ModalDialog,
	Sheet,
	Stack,
	Typography
} from '@mui/joy'
import DeleteIcon from '@mui/icons-material/Delete'

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
	const queryClient = useQueryClient()
	const query = useQuery({
		queryFn: async () => {
			const response = await fetch('/api/auth/me')
			const body = await response.json()
			return body?.token
		},
		queryKey: ['token'],
		retry: false
	})

	const deleteMutation = useMutation({
		mutationFn: async (pollId: string) => {
			const response = await fetch(`/api/polls/${pollId}`, {
				method: 'DELETE'
			})
			if (!response.ok) throw new Error('Failed to delete poll')
			return response.json()
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['userPolls'] })
		}
	})

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
						<Polls
							polls={polls}
							onDelete={deleteMutation.mutate}
							isDeleting={deleteMutation.isPending}
						/>
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

const Polls = ({
	polls,
	onDelete,
	isDeleting
}: {
	polls: Poll[] | undefined
	onDelete: (pollId: string) => void
	isDeleting: boolean
}) => {
	const [pollToDelete, setPollToDelete] = useState<string | null>(null)

	if (!polls?.length)
		return (
			<Alert color='neutral' variant='soft'>
				You haven't created any polls yet. Create your first poll to get
				started!
			</Alert>
		)

	return (
		<>
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
						(totalVoteCount, option) =>
							totalVoteCount + option.count,
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
										<Stack
											direction='row'
											spacing={1}
											alignItems='center'
										>
											<Typography level='body-sm'>
												Created: {createdDate}
											</Typography>
											<Button
												variant='soft'
												color='danger'
												size='md'
												onClick={() =>
													setPollToDelete(poll.id)
												}
												disabled={isDeleting}
											>
												<DeleteIcon />
											</Button>
										</Stack>
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

			<Modal
				open={!!pollToDelete}
				onClose={() => !isDeleting && setPollToDelete(null)}
			>
				<ModalDialog
					variant='outlined'
					role='alertdialog'
					aria-labelledby='delete-dialog-title'
					aria-describedby='delete-dialog-description'
				>
					<ModalClose disabled={isDeleting} />
					<Typography
						id='delete-dialog-title'
						level='h2'
						startDecorator={<DeleteIcon />}
						sx={{ color: 'danger.500' }}
					>
						Delete Poll
					</Typography>
					<Divider />
					<Typography id='delete-dialog-description' level='body-md'>
						Are you sure you want to delete this poll? This action
						cannot be undone.
					</Typography>
					<Box
						sx={{
							display: 'flex',
							gap: 1,
							justifyContent: 'flex-end',
							mt: 2
						}}
					>
						<Button
							variant='plain'
							color='neutral'
							onClick={() => setPollToDelete(null)}
							disabled={isDeleting}
						>
							Cancel
						</Button>
						<Button
							variant='solid'
							color='danger'
							onClick={() => {
								if (pollToDelete) {
									onDelete(pollToDelete)
								}
							}}
							loading={isDeleting}
							disabled={isDeleting}
						>
							Delete
						</Button>
					</Box>
				</ModalDialog>
			</Modal>
		</>
	)
}
