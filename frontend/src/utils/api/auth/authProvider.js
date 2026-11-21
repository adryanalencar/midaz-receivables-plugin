// @project
import { AUTH_PROVIDER } from '@/config';

// Mapping of auth types to dynamic imports
const authProviderMapping = {
  mock: () => import('@/utils/api/auth/mock/auth').then((mod) => mod.default),
  supabase: () => import('@/utils/api/auth/supabase/auth').then((mod) => mod.default),
  aws: () => import('@/utils/api/auth/aws/auth').then((mod) => mod.default)
};

// Dynamically loads and returns the auth provider based on AUTH_PROVIDER.
export async function authProvider() {
  return await authProviderMapping[AUTH_PROVIDER]();
}
