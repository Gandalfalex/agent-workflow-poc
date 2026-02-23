<script setup lang="ts">
import { Button } from "@/components/ui/button";

const props = defineProps<{
    identifier: string;
    password: string;
    loginError: string;
    loginBusy: boolean;
    isChecking: boolean;
    canLogin: boolean;
}>();

const emit = defineEmits<{
    (e: "update:identifier", value: string): void;
    (e: "update:password", value: string): void;
    (e: "submit"): void;
}>();

const onSubmit = () => emit("submit");
</script>

<template>
    <div
        data-testid="login.view"
        class="relative z-10 flex w-full flex-col gap-10 px-6 py-20 lg:flex-row"
    >
        <div class="flex-1 space-y-4">
            <p class="text-xs uppercase tracking-[0.3em] text-muted-foreground">
                Ops Console
            </p>
            <h1 class="text-4xl font-semibold">Sign in to the workspace</h1>
            <p class="text-sm text-muted-foreground">
                Use your Keycloak credentials to access the live board and
                webhooks.
            </p>
            <div
                class="rounded-2xl border border-border bg-card/70 px-4 py-3 text-xs text-muted-foreground"
            >
                Realm: ticketing · Client: myclient
            </div>
            <div
                class="rounded-2xl border border-border bg-card/70 px-4 py-3 text-xs text-muted-foreground"
            >
                Demo users: AdminUser / admin123 · NormalUser / user123
            </div>
        </div>

        <div
            class="w-full max-w-md rounded-3xl border border-border bg-card/80 p-6 shadow-sm"
        >
            <div class="space-y-2">
                <p
                    class="text-xs uppercase tracking-[0.3em] text-muted-foreground"
                >
                    Welcome back
                </p>
                <h2 class="text-2xl font-semibold">Team access</h2>
            </div>

            <form class="mt-6 space-y-4" @submit.prevent="onSubmit">
                <div>
                    <label class="text-xs font-semibold text-muted-foreground">
                        Email or username
                    </label>
                    <input
                        data-testid="login.identifier-input"
                        :value="props.identifier"
                        type="text"
                        autocomplete="username"
                        placeholder="AdminUser or ich@ich.ich"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        :disabled="props.isChecking"
                        @input="
                            emit(
                                'update:identifier',
                                ($event.target as HTMLInputElement).value,
                            )
                        "
                    />
                </div>
                <div>
                    <label class="text-xs font-semibold text-muted-foreground">
                        Password
                    </label>
                    <input
                        data-testid="login.password-input"
                        :value="props.password"
                        type="password"
                        autocomplete="current-password"
                        placeholder="••••••••"
                        class="mt-2 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        :disabled="props.isChecking"
                        @input="
                            emit(
                                'update:password',
                                ($event.target as HTMLInputElement).value,
                            )
                        "
                    />
                </div>
                <div
                    v-if="props.loginError"
                    data-testid="login.error"
                    class="rounded-2xl border border-border bg-secondary/60 px-3 py-2 text-xs"
                >
                    {{ props.loginError }}
                </div>
                <div
                    v-if="props.isChecking"
                    class="text-xs text-muted-foreground"
                >
                    Checking session...
                </div>
                <Button
                    data-testid="login.submit-button"
                    class="w-full"
                    type="submit"
                    :disabled="
                        !props.canLogin || props.loginBusy || props.isChecking
                    "
                >
                    {{
                        props.loginBusy || props.isChecking
                            ? "Signing in..."
                            : "Sign in"
                    }}
                </Button>
            </form>
        </div>
    </div>
</template>
