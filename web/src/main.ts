import { mount } from 'svelte'
import './styles/base.css'
import App from './App.svelte'
import { log } from './lib/log'

addEventListener('error', (e) => log('error', 'something broke', e.error ?? e.message))
addEventListener('unhandledrejection', (e) => log('error', 'a promise went unanswered', e.reason))

mount(App, { target: document.getElementById('app')! })
