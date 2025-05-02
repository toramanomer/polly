import { createContext, useContext, useEffect, useState } from 'react'

interface AuthContextType {
	token: string | null
	setToken: (token: string | null) => void
	isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
	const [token, setTokenState] = useState<string | null>(null)

	// Initialize token from httpOnly cookie
	useEffect(() => {
		// Check if we have a token in memory
		if (token) return

		// If not, try to get it from the server
		fetch('/api/auth/me', {
			credentials: 'include'
		})
			.then(res => res.json())
			.then(data => {
				if (data.token) setTokenState(data.token)
			})
			.catch(() => {
				// If we can't get the token, we're not authenticated
				setTokenState(null)
			})
	}, [token])

	const setToken = (newToken: string | null) => {
		setTokenState(newToken)
	}

	const value = {
		token,
		setToken,
		isAuthenticated: !!token
	}

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export const useAuth = () => {
	const context = useContext(AuthContext)
	if (!context) {
		throw new Error('useAuth must be used within an AuthProvider')
	}
	return context
}
