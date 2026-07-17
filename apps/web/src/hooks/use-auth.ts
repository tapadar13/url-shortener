"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { getSession, login, logout, register } from "@/lib/auth/browser-auth"

export const authSessionQueryKey = ["auth", "session"] as const

export function useAuthSession() {
  return useQuery({
    queryKey: authSessionQueryKey,
    queryFn: getSession,
    retry: false,
  })
}

export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: login,
    onSuccess: (user) => {
      queryClient.setQueryData(authSessionQueryKey, user)
    },
  })
}

export function useRegister() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: register,
    onSuccess: (user) => {
      queryClient.setQueryData(authSessionQueryKey, user)
    },
  })
}

export function useLogout() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: logout,
    onSettled: () => {
      queryClient.setQueryData(authSessionQueryKey, null)
    },
  })
}
