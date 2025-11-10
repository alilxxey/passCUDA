## Config example
```yaml
telegram:
  enabled: true
  token: "BOT_TOKEN"
  notify_chat_id: "CHAT_ID"
users: 
  - name: drunkbatya
    role: admin
    tg_user_id: USER_ID
    tg_allowed_chat_ids: []
door:
  relay_off_hold_time_s: 2
hardware:
  door_relay:
    chip: "gpiochip0"
    channel: 2
    active_level: true  # door is closed on this level
  open_btn:
    chip: "gpiochip0"
    channel: 3
    active_level: false  # button is released on this level
```
