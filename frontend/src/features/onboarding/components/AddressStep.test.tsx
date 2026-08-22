import { describe, expect, it, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderWithProviders } from '@/test/utils'
import { AddressStep } from './AddressStep'
import { ChainSelectStep } from './ChainSelectStep'

const SEED_PHRASE =
  'legal winner thank year wave sausage worth useful legal winner thank yellow'

describe('onboarding', () => {
  it('requires an explicit network choice', async () => {
    const onSelect = vi.fn()
    renderWithProviders(<ChainSelectStep selected={null} onSelect={onSelect} />)

    // Every MVP chain is offered; nothing is detected from the address.
    expect(screen.getByRole('button', { name: /Ethereum/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /TRON/ })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /Solana/ }))
    expect(onSelect).toHaveBeenCalledWith('solana')
  })

  it('rejects a malformed address before calling the backend', async () => {
    const onSubmit = vi.fn()
    renderWithProviders(
      <AddressStep
        chainId="ethereum"
        address="0xnope"
        onAddressChange={() => {}}
        onBack={() => {}}
        onSubmit={onSubmit}
        submitting={false}
        error={null}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: /analyze wallet/i }))

    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByRole('alert')).toHaveTextContent(
      /enter a valid wallet address/i,
    )
  })

  it('refuses anything that looks like a secret', () => {
    const onSubmit = vi.fn()
    renderWithProviders(
      <AddressStep
        chainId="ethereum"
        address={SEED_PHRASE}
        onAddressChange={() => {}}
        onBack={() => {}}
        onSubmit={onSubmit}
        submitting={false}
        error={null}
      />,
    )

    expect(screen.getByRole('alert')).toHaveTextContent(
      /private key or seed phrase/i,
    )
    expect(screen.getByText(/never asks for a private key/i)).toBeInTheDocument()
  })

  it('submits a valid public address', async () => {
    const onSubmit = vi.fn()
    const address = '0x71C7656EC7ab88b098defB751B7401B5f6d8976F'

    renderWithProviders(
      <AddressStep
        chainId="ethereum"
        address={address}
        onAddressChange={() => {}}
        onBack={() => {}}
        onSubmit={onSubmit}
        submitting={false}
        error={null}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: /analyze wallet/i }))
    expect(onSubmit).toHaveBeenCalledWith(address)
  })
})
