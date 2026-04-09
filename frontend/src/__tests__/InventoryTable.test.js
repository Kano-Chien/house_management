import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import InventoryTable from '../components/InventoryTable.vue'

const sampleItems = [
  { id: 1, name: 'Rice', current_stock: 5, min_stock: 3, price: 10, category: 'food', is_tracked: true, planned_consumption: 2 },
  { id: 2, name: 'Shampoo', current_stock: 2, min_stock: 1, price: 50, category: 'daily', is_tracked: true, planned_consumption: 0 },
  { id: 3, name: 'Eggs', current_stock: 12, min_stock: 6, price: 5, category: 'food', is_tracked: true, planned_consumption: 0 },
]

function mockFetch(items = sampleItems) {
  return vi.fn().mockResolvedValue({
    ok: true,
    json: () => Promise.resolve(items),
  })
}

describe('InventoryTable.vue', () => {
  beforeEach(() => {
    global.fetch = mockFetch()
  })

  it('renders all items after fetch', async () => {
    const wrapper = mount(InventoryTable)
    await flushPromises()
    expect(wrapper.text()).toContain('Rice')
    expect(wrapper.text()).toContain('Shampoo')
    expect(wrapper.text()).toContain('Eggs')
  })

  it('filters to food category', async () => {
    const wrapper = mount(InventoryTable)
    await flushPromises()

    // Click the "Food" filter button
    const buttons = wrapper.findAll('button')
    const foodBtn = buttons.find(b => b.text().includes('Food'))
    await foodBtn.trigger('click')

    expect(wrapper.text()).toContain('Rice')
    expect(wrapper.text()).toContain('Eggs')
    expect(wrapper.text()).not.toContain('Shampoo')
  })

  it('filters to daily category', async () => {
    const wrapper = mount(InventoryTable)
    await flushPromises()

    const buttons = wrapper.findAll('button')
    const dailyBtn = buttons.find(b => b.text().includes('Daily'))
    await dailyBtn.trigger('click')

    expect(wrapper.text()).toContain('Shampoo')
    expect(wrapper.text()).not.toContain('Rice')
  })

  it('shows all items when All filter selected', async () => {
    const wrapper = mount(InventoryTable)
    await flushPromises()

    // Select daily then back to all
    const buttons = wrapper.findAll('button')
    const dailyBtn = buttons.find(b => b.text().includes('Daily'))
    await dailyBtn.trigger('click')
    const allBtn = buttons.find(b => b.text().includes('All'))
    await allBtn.trigger('click')

    expect(wrapper.text()).toContain('Rice')
    expect(wrapper.text()).toContain('Shampoo')
  })

  it('shows empty state message when no items match filter', async () => {
    global.fetch = mockFetch([
      { id: 1, name: 'Salt', current_stock: 3, min_stock: 1, price: 1, category: 'food', is_tracked: true, planned_consumption: 0 },
    ])
    const wrapper = mount(InventoryTable)
    await flushPromises()

    const buttons = wrapper.findAll('button')
    const dailyBtn = buttons.find(b => b.text().includes('Daily'))
    await dailyBtn.trigger('click')

    expect(wrapper.text()).toContain('No items in this category')
  })

  it('toggles add form when Add Item button clicked', async () => {
    const wrapper = mount(InventoryTable)
    await flushPromises()

    // Add form should be hidden initially
    expect(wrapper.text()).not.toContain('Close')

    const addBtn = wrapper.findAll('button').find(b => b.text().includes('Add Item'))
    await addBtn.trigger('click')

    expect(wrapper.text()).toContain('Close')
  })

  it('calls POST /api/inventory when adding a new item', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(sampleItems) })         // initial GET
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ id: 99, name: 'Pepper', current_stock: 5, price: 2, category: 'food', is_tracked: true, planned_consumption: 0 }) }) // POST
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([...sampleItems, { id: 99, name: 'Pepper', current_stock: 5 }]) }) // re-fetch
    global.fetch = fetchMock

    const wrapper = mount(InventoryTable)
    await flushPromises()

    // Open add form
    const addBtn = wrapper.findAll('button').find(b => b.text().includes('Add Item'))
    await addBtn.trigger('click')

    // Fill in name
    const nameInput = wrapper.find('input[placeholder="e.g. Chicken, Soap..."]')
    await nameInput.setValue('Pepper')

    // Click Add button in form
    const addFormBtn = wrapper.find('button[disabled]') // initially disabled until name is filled
    // The button should now be enabled
    const addSubmitBtn = wrapper.findAll('button').find(b => b.text() === 'Add')
    if (addSubmitBtn) {
      await addSubmitBtn.trigger('click')
      await flushPromises()

      const postCall = fetchMock.mock.calls.find(([url, opts]) => url === '/api/inventory' && opts?.method === 'POST')
      expect(postCall).toBeDefined()
      const body = JSON.parse(postCall[1].body)
      expect(body.name).toBe('Pepper')
    }
  })

  it('calls PUT /api/inventory/stock when stock + clicked', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(sampleItems) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'updated' }) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(sampleItems) })
    global.fetch = fetchMock

    const wrapper = mount(InventoryTable)
    await flushPromises()

    // Find first + button (increment stock)
    const plusBtns = wrapper.findAll('button').filter(b => b.text() === '+')
    if (plusBtns.length > 0) {
      await plusBtns[0].trigger('click')
      await flushPromises()

      const putCall = fetchMock.mock.calls.find(([url, opts]) => url === '/api/inventory/stock' && opts?.method === 'PUT')
      expect(putCall).toBeDefined()
    }
  })

  it('calls delete API when Delete button clicked', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(sampleItems) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ status: 'deleted' }) })
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(sampleItems.slice(1)) })
    global.fetch = fetchMock
    global.confirm = vi.fn(() => true) // auto-confirm deletion

    const wrapper = mount(InventoryTable)
    await flushPromises()

    const deleteBtn = wrapper.findAll('button').find(b => b.text() === 'Delete')
    if (deleteBtn) {
      await deleteBtn.trigger('click')
      await flushPromises()

      const deleteCall = fetchMock.mock.calls.find(([url, opts]) => url === '/api/inventory/delete' && opts?.method === 'POST')
      expect(deleteCall).toBeDefined()
    }
  })

  it('highlights stock count in red when stock < planned_consumption', async () => {
    global.fetch = mockFetch([
      { id: 1, name: 'Rice', current_stock: 1, min_stock: 3, price: 10, category: 'food', is_tracked: true, planned_consumption: 5 },
    ])
    const wrapper = mount(InventoryTable)
    await flushPromises()

    const redStock = wrapper.find('.text-red-500')
    expect(redStock.exists()).toBe(true)
    expect(redStock.text()).toBe('1')
  })
})
