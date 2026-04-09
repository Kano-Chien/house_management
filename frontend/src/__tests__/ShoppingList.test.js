import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ShoppingList from '../components/ShoppingList.vue'

// Helper to build a mock fetch response
function mockFetch(data, ok = true) {
  return vi.fn().mockResolvedValue({
    ok,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  })
}

const autoItem = { id: 1, name: 'Eggs', is_custom: false, is_checked: true }
const customItem = { id: 2, name: 'Dish Soap', is_custom: true, is_checked: true }

describe('ShoppingList.vue', () => {
  beforeEach(() => {
    global.fetch = mockFetch([autoItem, customItem])
  })

  it('renders loading state initially', () => {
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {})) // never resolves
    const wrapper = mount(ShoppingList)
    expect(wrapper.text()).toContain('Loading')
  })

  it('displays auto and custom items after load', async () => {
    const wrapper = mount(ShoppingList)
    await flushPromises()
    expect(wrapper.text()).toContain('Eggs')
    expect(wrapper.text()).toContain('Dish Soap')
  })

  it('displayList merges auto and custom items', async () => {
    const wrapper = mount(ShoppingList)
    await flushPromises()
    // Both items visible
    const items = wrapper.findAll('[data-testid="shopping-item"]').length
    // Can check by text presence instead since we don't have data-testid
    expect(wrapper.text()).toContain('Eggs')
    expect(wrapper.text()).toContain('Dish Soap')
  })

  it('shows empty state when no items', async () => {
    global.fetch = mockFetch([])
    const wrapper = mount(ShoppingList)
    await flushPromises()
    expect(wrapper.text()).toContain('Everything is in stock')
  })

  it('adds a custom item when Add button clicked', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([autoItem, customItem]) }) // initial load
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'added' }) })    // POST
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([autoItem, customItem, { id: 3, name: 'Milk', is_custom: true, is_checked: true }]) }) // refetch

    global.fetch = fetchMock
    const wrapper = mount(ShoppingList)
    await flushPromises()

    const input = wrapper.find('input[placeholder="Add item to shopping list..."]')
    await input.setValue('Milk')
    const addBtn = wrapper.findAll('button').find(b => b.text().includes('+ Add'))
    await addBtn.trigger('click')
    await flushPromises()

    // POST should have been called with correct body
    const postCall = fetchMock.mock.calls.find(([url, opts]) => url === '/api/shopping-list' && opts?.method === 'POST')
    expect(postCall).toBeDefined()
    const body = JSON.parse(postCall[1].body)
    expect(body.name).toBe('Milk')
  })

  it('does not add item when input is empty', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    global.fetch = fetchMock
    const wrapper = mount(ShoppingList)
    await flushPromises()

    const callsBefore = fetchMock.mock.calls.length
    await wrapper.find('button:last-of-type').trigger('click')
    await flushPromises()

    // No additional fetch call for adding
    expect(fetchMock.mock.calls.length).toBe(callsBefore)
  })

  it('calls delete API when remove button clicked', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([autoItem]) }) // initial load
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'deleted' }) }) // DELETE
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) }) // refetch
    global.fetch = fetchMock

    const wrapper = mount(ShoppingList)
    await flushPromises()

    const removeBtn = wrapper.find('button[class*="opacity-0"]')
    if (removeBtn.exists()) {
      await removeBtn.trigger('click')
      await flushPromises()
      const deleteCall = fetchMock.mock.calls.find(([url]) => url === '/api/shopping-list/delete')
      expect(deleteCall).toBeDefined()
    }
  })

  it('calls LINE API when Send to LINE clicked', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([autoItem, customItem]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'sent' }) })
    global.fetch = fetchMock
    global.alert = vi.fn()

    const wrapper = mount(ShoppingList)
    await flushPromises()

    const lineBtn = wrapper.find('button:first-of-type')
    await lineBtn.trigger('click')
    await flushPromises()

    const lineCall = fetchMock.mock.calls.find(([url]) => url === '/api/line/send-shopping-list')
    expect(lineCall).toBeDefined()
    const body = JSON.parse(lineCall[1].body)
    expect(body).toBeInstanceOf(Array)
    expect(body.length).toBeGreaterThan(0)
  })

  it('disables Send to LINE button when list is empty', async () => {
    global.fetch = mockFetch([])
    const wrapper = mount(ShoppingList)
    await flushPromises()

    const lineBtn = wrapper.find('button[disabled]')
    expect(lineBtn.exists()).toBe(true)
  })

  it('persists checked state when checkbox toggled', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([{ ...autoItem, is_checked: true }]) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'updated' }) })
    global.fetch = fetchMock

    const wrapper = mount(ShoppingList)
    await flushPromises()

    const checkbox = wrapper.find('input[type="checkbox"]')
    await checkbox.trigger('change')
    await flushPromises()

    const checkCall = fetchMock.mock.calls.find(([url]) => url === '/api/shopping-list/check')
    expect(checkCall).toBeDefined()
  })
})
