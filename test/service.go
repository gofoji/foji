package test

import "context"

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

const ErrNotImplemented = ServiceError("not implemented")

// New creates a new service instance.
func New() *Service {
	return &Service{}
}

// Service implements all business logic for test.
type Service struct {
}

// FindPets Returns all pets from the system that the user has access to Nam sed condimentum est.
// Maecenas tempor sagittis sapien, nec rhoncus sem sagittis sit amet. Aenean at gravida augue,
// ac iaculis sem. Curabitur odio lorem, ornare eget elementum nec, cursus id lectus. Duis mi
// turpis, pulvinar ac eros ac, tincidunt varius justo. In hac habitasse platea dictumst.
// Integer at adipiscing ante, a sagittis ligula. Aenean pharetra tempor ante molestie
// imperdiet. Vivamus id aliquam diam. Cras quis velit non tortor eleifend sagittis. Praesent at
// enim pharetra urna volutpat venenatis eget eget mauris. In eleifend fermentum facilisis.
// Praesent enim enim, gravida ac sodales sed, placerat id erat. Suspendisse lacus dolor,
// consectetur non augue vel, vehicula interdum libero. Morbi euismod sagittis libero sed
// lacinia.
//
// Sed tempus felis lobortis leo pulvinar rutrum. Nam mattis velit nisl, eu condimentum
// ligula luctus nec. Phasellus semper velit eget aliquet faucibus. In a mattis elit. Phasellus
// vel urna viverra, condimentum lorem id, rhoncus nibh. Ut pellentesque posuere elementum. Sed
// a varius odio. Morbi rhoncus ligula libero, vel eleifend nunc tristique vitae. Fusce et sem
// dui. Aenean nec scelerisque tortor. Fusce malesuada accumsan magna vel tempus. Quisque
// mollis felis eu dolor tristique, sit amet auctor felis gravida. Sed libero lorem, molestie sed
// nisl in, accumsan tempor nisi. Fusce sollicitudin massa ut lacinia mattis. Sed vel eleifend
// lorem. Pellentesque vitae felis pretium, pulvinar elit eu, euismod sapien.
//
//.
func (s *Service) FindPets(ctx context.Context, tags []string, limit *int32) ([]Pet, error) {
	return nil, ErrNotImplemented
}

// AddPet Creates a new pet in the store. Duplicates are allowed.
func (s *Service) AddPet(ctx context.Context, user *User, newPet NewPet) (*Pet, error) {
	out := Pet{
		NewPet: NewPet{
			Name: "MyName",
			Tag:  user.ID,
		},
		ID:     101,
	}
	return &out, nil
}

// DeletePet deletes a single pet based on the ID supplied.
func (s *Service) DeletePet(ctx context.Context, id int64) error {
	return ErrNotImplemented
}

// FindPetByID Returns a user based on a single ID, if the user does not have access to the pet.
func (s *Service) FindPetByID(ctx context.Context, id int64) (*Pet, error) {
	return nil, ErrNotImplemented
}
