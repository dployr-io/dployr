// src/index.ts
import { Hono } from 'hono'
import { cors } from 'hono/cors'

type Bindings = {
  ZEPTO_API_KEY: string
}

const app = new Hono<{ Bindings: Bindings }>()

app.use('/api/*', cors())

// Email API endpoint
app.post('/api/send-email', async (c) => {
  try {
    const { to, subject, body, name } = await c.req.json()
    
    if (!to || !subject || !body) {
      return c.json({ error: 'Missing required fields' }, 400)
    }

    const emailPayload = {
      to: [{ 
        email_address: { 
          address: to, 
          name: name || to 
        } 
      }],
      from: { 
        address: 'noreply@zeipo.ai'
      },
      subject,
      htmlbody: body
    }

    const response = await fetch('https://api.zeptomail.com/v1.1/email', {
      method: 'POST',
      headers: {
        'Authorization': `Zoho-enczapikey ${c.env.ZEPTO_API_KEY}`,
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(emailPayload)
    })

    if (!response.ok) {
      const error = await response.text()
      console.error('Zepto API error:', error)
      return c.json({ error: 'Failed to send email' }, 500)
    }

    return c.json({ message: 'Email sent successfully' })
  } catch (error) {
    console.error('Email error:', error)
    return c.json({ error: 'Internal server error' }, 500)
  }
})

// Health check
app.get('/api/health', (c) => {
  return c.json({ status: 'ok', timestamp: new Date().toISOString() })
})


export default app