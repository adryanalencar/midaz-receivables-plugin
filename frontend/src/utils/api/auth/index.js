// @project
import { authProvider } from './authProvider';
import { AUTH_USER_KEY } from '@/config';
import { attempt } from '@/utils/attempt';

export async function login(formData) {
  const authHandler = await authProvider();

  if (!authHandler.login) {
    return { data: null, error: 'Login not supported by current provider' };
  }

  return attempt(authHandler.login(formData));
}

export async function signUp(formData) {
  const authHandler = await authProvider();

  if (!authHandler.signUp) {
    return { data: null, error: 'Sign-up not supported by current provider' };
  }

  return attempt(authHandler.signUp(formData));
}

export async function getUser() {
  const authHandler = await authProvider();

  if (!authHandler.getUser) {
    return { data: null, error: 'Get user not supported by current provider' };
  }

  return attempt(authHandler.getUser());
}

export async function forgotPassword(formData) {
  const authHandler = await authProvider();

  if (!authHandler.forgotPassword) {
    return { data: null, error: 'Forgot password not supported by current provider' };
  }

  return attempt(authHandler.forgotPassword(formData));
}

export async function resetPassword(formData) {
  const authHandler = await authProvider();

  if (!authHandler.resetPassword) {
    return { data: null, error: 'Reset password not supported by current provider' };
  }

  return attempt(authHandler.resetPassword(formData));
}

export async function verifyOtp(formData) {
  const authHandler = await authProvider();

  if (!authHandler.verifyOtp) {
    return { data: null, error: 'Verify OTP not supported by current provider' };
  }

  return attempt(authHandler.verifyOtp(formData));
}

export async function resendOtp(formData) {
  const authHandler = await authProvider();

  if (!authHandler.resend) {
    return { data: null, error: 'Resend OTP not supported by current provider' };
  }

  return attempt(authHandler.resend(formData));
}

export async function logout() {
  const authHandler = await authProvider();

  if (!authHandler.signOut) {
    return { data: null, error: 'SignOut functionality not available' };
  }

  await authHandler.signOut();
  localStorage.removeItem(AUTH_USER_KEY);
  window.location.pathname = '/login';
  return { data: { message: 'Loggedout' }, error: null };
}
