// Mock EC2 Client
// Copyright (C) 2017+ Jonathan Gold and the project contributors
// Written by Jonathan Gold <info@jonathangold.ca> and the project
// contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ec2mock

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *MockEc2Client) StartInstances(in *ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error) {
	for _, instanceID := range in.InstanceIds {
		for _, res := range c.Reservations {
			for _, inst := range res.Instances {
				if inst.InstanceId != instanceID {
					continue
				}
				switch aws.StringValue(inst.State.Name) {
				case ec2.InstanceStateNameStopped:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNamePending)
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameRunning)
					}()
				case ec2.InstanceStateNamePending:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameRunning)
					}()
				case ec2.InstanceStateNameRunning:
				default:
					return nil, awserr.New(
						AwsErrIncorrectInstanceState,
						"",
						fmt.Errorf(AwsErrIncorrectInstanceState),
					)
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}
