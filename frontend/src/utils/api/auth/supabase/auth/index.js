// @project
import { AUTH_USER_KEY } from '@/config';
import { AuthRole } from '@/enum';
import { createSupabaseClient } from '@/utils/auth-client/supabase';

const supabase = createSupabaseClient();

/***************************  SUPABASE - LOGIN  ***************************/

export async function login(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { data, error } = await supabase.auth.signInWithPassword({
        email: formData.email,
        password: formData.password
      });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({
        id: data.user?.id,
        email: data.user?.email || '',
        access_token: data.session?.access_token
      });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - GET USER  ***************************/

export async function getUser() {
  return new Promise(async (resolve, reject) => {
    try {
      const storedValue = typeof window !== 'undefined' ? localStorage.getItem(AUTH_USER_KEY) : null;
      const parsedValue = storedValue && JSON.parse(storedValue);
      const token = parsedValue?.access_token;

      if (!token) {
        reject(new Error('Token not found'));
        return;
      }

      const { data, error } = await supabase.auth.getUser(token);

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({
        id: data.user.id,
        email: data.user.email || '',
        role: data.user.user_metadata.role,
        contact: data.user.user_metadata.contact,
        dialcode: data.user.user_metadata.dialcode,
        firstname: data.user.user_metadata.firstname,
        lastname: data.user.user_metadata.lastname
      });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - SIGN UP  ***************************/

export async function signUp(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { error } = await supabase.auth.signUp({
        email: formData.email,
        password: formData.password,
        options: {
          data: {
            firstname: formData.firstname,
            lastname: formData.lastname,
            dialcode: formData.dialcode,
            contact: formData.contact,
            role: AuthRole.USER
          }
        }
      });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - VERIFY OTP  ***************************/

export async function verifyOtp(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { error } = await supabase.auth.verifyOtp({
        email: formData.email,
        token: formData.otp,
        type: formData.type
      });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - RESEND OTP  ***************************/

export async function resend(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { error } = await supabase.auth.resend({
        email: formData.email,
        type: formData.type
      });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - FORGOT PASSWORD  ***************************/

export async function forgotPassword(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { error } = await supabase.auth.resetPasswordForEmail(formData.email, {
        redirectTo: formData.redirectTo
      });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - RESET PASSWORD  ***************************/

export async function resetPassword(formData) {
  return new Promise(async (resolve, reject) => {
    try {
      const { error: sessionError } = await supabase.auth.setSession({
        access_token: formData.accessToken,
        refresh_token: formData.refreshToken
      });

      if (sessionError) {
        reject(new Error('Unauthorized'));
        return;
      }

      const { error } = await supabase.auth.updateUser({ password: formData.password });

      if (error) {
        reject(new Error(error.message));
        return;
      }

      await supabase.auth.signOut();

      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

/***************************  SUPABASE - SIGN OUT  ***************************/

export async function signOut() {
  return new Promise(async (resolve, reject) => {
    try {
      await supabase.auth.signOut();
      resolve({ status: 200 });
    } catch {
      reject(new Error('Server error'));
    }
  });
}

// Export as a single object for easy import
const supabaseAuth = { login, getUser, signUp, verifyOtp, resend, forgotPassword, resetPassword, signOut };

export default supabaseAuth;
