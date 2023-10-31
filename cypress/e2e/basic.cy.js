describe('empty spec', () => {
  beforeEach(() => {
    cy.visit('/')
  })
  it('displays the title', () => {
    cy.get('h1')
    .contains('The Gift of Nothing');
  })
})