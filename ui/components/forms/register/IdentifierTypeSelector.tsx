'use client'

import { useFormContext } from 'react-hook-form'

interface IdentifierTypeSelectorProps {
  errors: { message?: string } | undefined
}

export function IdentifierTypeSelector({ errors }: IdentifierTypeSelectorProps) {
  const { register, watch, setValue } = useFormContext()
  const identifierType = watch('identifier_type')

  return (
    <div className="space-y-1.5">
      <label className="text-xs font-medium text-slate-700 dark:text-slate-200">
        注册方式
      </label>
      <div className="relative">
        <input
          type="hidden"
          {...register('identifier_type')}
          value={identifierType}
        />
        <div
          className="relative flex w-full rounded-lg bg-slate-200 dark:bg-slate-700 p-1"
          role="radiogroup"
          aria-label="注册方式"
        >
          <div
            className={`absolute top-1 bottom-1 w-1/2 rounded-md bg-gradient-to-r from-blue-600 to-indigo-600 transition-transform duration-300 ease-in-out ${
              identifierType === 'phone' ? 'translate-x-full' : 'translate-x-0'
            }`}
          />

          <button
            type="button"
            role="radio"
            aria-checked={identifierType === 'email'}
            onClick={() => setValue('identifier_type', 'email')}
            className={`relative z-10 flex-1 py-2 text-xs font-medium rounded-md transition-colors duration-300 ${
              identifierType === 'email'
                ? 'text-white'
                : 'text-slate-600 dark:text-slate-300'
            }`}
          >
            邮箱注册
          </button>

          <button
            type="button"
            role="radio"
            aria-checked={identifierType === 'phone'}
            onClick={() => setValue('identifier_type', 'phone')}
            className={`relative z-10 flex-1 py-2 text-xs font-medium rounded-md transition-colors duration-300 ${
              identifierType === 'phone'
                ? 'text-white'
                : 'text-slate-600 dark:text-slate-300'
            }`}
          >
            手机号注册
          </button>
        </div>
      </div>
      {errors && (
        <p className="mt-1.5 text-xs text-red-600 flex items-center">
          <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
          {errors.message}
        </p>
      )}
    </div>
  )
}
